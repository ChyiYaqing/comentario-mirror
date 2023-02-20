package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/markbates/goth"
	"github.com/pkg/errors"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
	"net/url"
	"time"
)

type ssoPayload struct {
	Domain string `json:"domain"`
	Token  string `json:"token"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Link   string `json:"link"`
	Photo  string `json:"photo"`
}

// oauthSessions stores initiated OAuth (federated authentication) sessions
var oauthSessions = map[string]string{}

// OauthInit initiates a federated authentication process
func OauthInit(params operations.OauthInitParams) middleware.Responder {
	// Get the registered provider instance by its name (coming from the path parameter)
	provider, err := goth.GetProvider(params.Provider)
	if err != nil {
		return operations.NewGenericBadRequest().WithPayload(&operations.GenericBadRequestBody{
			Details: fmt.Sprintf("%s (%s)", util.ErrorOAuthNotConfigured.Error(), params.Provider),
		})
	}

	// Verify the provided commenter token
	if _, err = commenterGetByCommenterToken(models.CommenterHexID(params.CommenterToken)); err != nil && err != util.ErrorNoSuchToken {
		return oauthFailure(err)
	}

	// Initiate an authentication session
	sess, err := provider.BeginAuth(params.CommenterToken)
	if err != nil {
		logger.Warningf("OauthInit(): provider.BeginAuth() failed: %v", err)
		return operations.NewGenericInternalServerError()
	}

	// Fetch the redirection URL
	sessURL, err := sess.GetAuthURL()
	if err != nil {
		logger.Warningf("OauthInit(): sess.GetAuthURL() failed: %v", err)
		return operations.NewGenericInternalServerError()
	}

	// Store the session in memory, to verify it later
	sessID, _ := util.RandomHex(32)
	oauthSessions[sessID] = sess.Marshal()

	// Succeeded: redirect the user to the federated identity provider, setting the state cookie
	return NewCookieResponder(operations.NewOauthInitTemporaryRedirect().WithLocation(sessURL)).
		WithCookie(
			util.AuthSessionCookieName,
			sessID,
			"/",
			time.Hour, // One hour must be sufficient to complete authentication
			true,
			http.SameSiteLaxMode)
}

func OauthCallback(params operations.OauthCallbackParams) middleware.Responder {
	// Get the registered provider instance by its name (coming from the path parameter)
	provider, err := goth.GetProvider(params.Provider)
	if err != nil {
		logger.Debugf("Failed to fetch provider '%s': %v", params.Provider, err)
		return oauthFailure(fmt.Errorf("unknown provider: %s", params.Provider))
	}

	// Obtain the auth session ID from the cookie
	var sess goth.Session
	if cookie, err := params.HTTPRequest.Cookie(util.AuthSessionCookieName); err != nil {
		logger.Debugf("Auth session cookie error: %v", err)
		return oauthFailure(errors.New("auth session cookie missing"))

		// Find and delete the session
	} else if sessData, ok := oauthSessions[cookie.Value]; !ok {
		logger.Debugf("No auth session found with ID=%v: %v", cookie.Value, err)
		return oauthFailure(errors.New("auth session not found"))

	} else {
		// Delete the locally stored session
		delete(oauthSessions, cookie.Value)

		// Recover the original provider session
		if sess, err = provider.UnmarshalSession(sessData); err != nil {
			logger.Debugf("provider.UnmarshalSession() failed: %v", err)
			return oauthFailure(errors.New("auth session unmarshalling"))
		}
	}

	// Validate the session state
	if err := validateAuthSessionState(sess, params.HTTPRequest); err != nil {
		return oauthFailure(err)
	}

	// Obtain the tokens
	reqParams := params.HTTPRequest.URL.Query()
	_, err = sess.Authorize(provider, reqParams)
	if err != nil {
		logger.Debugf("sess.Authorize() failed: %v", err)
		return oauthFailure(errors.New("auth session unauthorised"))
	}

	// Fetch the federated user
	fedUser, err := provider.FetchUser(sess)
	if err != nil {
		logger.Debugf("provider.FetchUser() failed: %v", err)
		return oauthFailure(errors.New("fetching user"))
	}

	// Validate the returned user
	// -- UserID
	if fedUser.UserID == "" {
		return oauthFailure(errors.New("user ID missing"))
	}
	// -- Email
	if fedUser.Email == "" {
		return oauthFailure(errors.New("user email missing"))
	}
	// -- Name
	if fedUser.Name == "" {
		return oauthFailure(errors.New("user name missing"))
	}

	// Try to find the corresponding existing user in the database
	commenterToken := models.CommenterHexID(reqParams.Get("state"))
	if _, err = commenterGetByCommenterToken(commenterToken); err != nil && err != util.ErrorNoSuchToken {
		return oauthFailure(err)
	}

	commenter, err := commenterGetByEmail(params.Provider, strfmt.Email(fedUser.Email))
	if err != nil && err != util.ErrorNoSuchCommenter {
		return oauthFailure(err)
	}

	var commenterHex models.CommenterHexID
	avatar := fedUser.AvatarURL
	if avatar == "" {
		// TODO get rid of this crap
		avatar = "undefined"
	}
	// No such commenter yet: it's a signup
	if err == util.ErrorNoSuchCommenter {
		// Create a new commenter
		if commenterHex, err = commenterNew(strfmt.Email(fedUser.Email), fedUser.Name, "undefined", avatar, params.Provider, ""); err != nil {
			return oauthFailure(err)
		}

		// Commenter already exists: it's a login
	} else {
		// Update commenter's details
		if err = commenterUpdate(commenter.CommenterHex, strfmt.Email(fedUser.Email), fedUser.Name, "undefined", avatar, params.Provider); err != nil {
			logger.Warningf("cannot update commenter: %s", err)
			// Don't exit as we still can proceed
		}
		commenterHex = commenter.CommenterHex
	}

	// Register a commenter session
	if err := commenterSessionUpdate(models.HexID(commenterToken), commenterHex); err != nil {
		return oauthFailure(err)
	}

	// Succeeded: close the parent window, removing the auth session cookie
	return NewCookieResponder(closeParentWindowResponse()).WithoutCookie(util.AuthSessionCookieName, "/")
}

func OauthSsoCallback(params operations.OauthSsoCallbackParams) middleware.Responder {
	payloadBytes, err := hex.DecodeString(params.Payload)
	if err != nil {
		return oauthFailure(fmt.Errorf("payload: invalid hex encoding: %s", err.Error()))
	}

	signatureBytes, err := hex.DecodeString(params.Hmac)
	if err != nil {
		return oauthFailure(fmt.Errorf("HMAC signature: invalid hex encoding: %s", err.Error()))
	}

	payload := ssoPayload{}
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return oauthFailure(fmt.Errorf("payload: failed to unmarshal: %s", err.Error()))
	}

	if payload.Token == "" || payload.Email == "" || payload.Name == "" {
		return oauthFailure(util.ErrorMissingField)
	}

	if payload.Link == "" {
		payload.Link = "undefined"
	}

	if payload.Photo == "" {
		payload.Photo = "undefined"
	}

	domain, commenterToken, err := ssoTokenExtract(payload.Token)
	if err != nil {
		return oauthFailure(err)
	}

	d, err := domainGet(domain)
	if err != nil {
		if err == util.ErrorNoSuchDomain {
			return oauthFailure(err)
		}
		logger.Errorf("cannot get domain for SSO: %v", err)
		return oauthFailure(util.ErrorInternal)
	}

	if d.SsoSecret == "" || d.SsoURL == "" {
		return oauthFailure(util.ErrorMissingConfig)
	}

	key, err := hex.DecodeString(d.SsoSecret)
	if err != nil {
		logger.Errorf("cannot decode SSO secret as hex: %v", err)
		return oauthFailure(err)
	}

	h := hmac.New(sha256.New, key)
	h.Write(payloadBytes)
	expectedSignatureBytes := h.Sum(nil)
	if !hmac.Equal(expectedSignatureBytes, signatureBytes) {
		return oauthFailure(fmt.Errorf("HMAC signature verification failed"))
	}

	_, err = commenterGetByCommenterToken(commenterToken)
	if err != nil && err != util.ErrorNoSuchToken {
		return oauthFailure(err)
	}

	c, err := commenterGetByEmail("sso:"+domain, strfmt.Email(payload.Email))
	if err != nil && err != util.ErrorNoSuchCommenter {
		return oauthFailure(err)
	}

	var commenterHex models.CommenterHexID
	if err == util.ErrorNoSuchCommenter {
		if commenterHex, err = commenterNew(strfmt.Email(payload.Email), payload.Name, payload.Link, payload.Photo, "sso:"+domain, ""); err != nil {
			return oauthFailure(err)
		}
	} else {
		if err = commenterUpdate(c.CommenterHex, strfmt.Email(payload.Email), payload.Name, payload.Link, payload.Photo, "sso:"+domain); err != nil {
			logger.Warningf("cannot update commenter: %s", err)
			// not a serious enough to exit with an error
		}

		commenterHex = c.CommenterHex
	}

	if err = commenterSessionUpdate(models.HexID(commenterToken), commenterHex); err != nil {
		return oauthFailure(err)
	}

	// Succeeded: close the parent window
	return closeParentWindowResponse()
}

func OauthSsoRedirect(params operations.OauthSsoRedirectParams) middleware.Responder {
	domainURL, err := util.ParseAbsoluteURL(params.HTTPRequest.Header.Get("Referer"))
	if err != nil {
		return oauthFailure(err)
	}
	domainName := domainURL.Host

	if _, err = commenterGetByCommenterToken(models.CommenterHexID(params.CommenterToken)); err != nil && err != util.ErrorNoSuchToken {
		return oauthFailure(err)
	}

	d, err := domainGet(domainName)
	if err != nil {
		return oauthFailure(util.ErrorNoSuchDomain)
	}
	if !d.Idps["sso"] {
		return oauthFailure(fmt.Errorf("SSO not configured for %s", domainName))
	}
	if d.SsoSecret == "" || d.SsoURL == "" {
		return oauthFailure(util.ErrorMissingConfig)
	}

	key, err := hex.DecodeString(d.SsoSecret)
	if err != nil {
		logger.Errorf("cannot decode SSO secret as hex: %v", err)
		return oauthFailure(err)
	}

	token, err := ssoTokenNew(domainName, params.CommenterToken)
	if err != nil {
		return oauthFailure(err)
	}

	tokenBytes, err := hex.DecodeString(token)
	if err != nil {
		logger.Errorf("cannot decode hex token: %v", err)
		return oauthFailure(util.ErrorInternal)
	}

	h := hmac.New(sha256.New, key)
	h.Write(tokenBytes)
	signature := hex.EncodeToString(h.Sum(nil))

	ssoURL, err := util.ParseAbsoluteURL(d.SsoURL)
	if err != nil {
		// this should really not be happening; we're checking if the passed URL is valid at domain update
		logger.Errorf("cannot parse URL: %v", err)
		return oauthFailure(util.ErrorInternal)
	}

	q := ssoURL.Query()
	q.Set("token", token)
	q.Set("hmac", signature)
	ssoURL.RawQuery = q.Encode()

	// Succeeded
	return operations.NewOauthSsoRedirectTemporaryRedirect().WithLocation(ssoURL.String())
}

// oauthFailure returns a generic "Unauthorized" responder, with the error message in the details. Also wipes out any
// auth session cookie
func oauthFailure(err error) middleware.Responder {
	return NewCookieResponder(
		operations.NewOauthInitUnauthorized().
			WithPayload(fmt.Sprintf(
				`<html lang="en">
				<head>
					<title>401 Unauthorized</title>
				</head>
				<body>
					<h1>Unauthorized</h1>
					<p>OAuth authentication failed with the error: %s</p>
				</body>
				</html>`,
				err.Error()))).
		WithoutCookie(util.AuthSessionCookieName, "/")
}

func ssoTokenExtract(token string) (string, models.CommenterHexID, error) {
	statement := "select domain, commenterToken from ssoTokens where token = $1;"
	row := svc.DB.QueryRow(statement, token)

	var domain string
	var commenterToken models.CommenterHexID
	if err := row.Scan(&domain, &commenterToken); err != nil {
		return "", "", util.ErrorNoSuchToken
	}

	statement = `
		delete from ssoTokens
		where token = $1;
	`
	if _, err := svc.DB.Exec(statement, token); err != nil {
		logger.Errorf("cannot delete SSO token after usage: %v", err)
		return "", "", util.ErrorInternal
	}

	return domain, commenterToken, nil
}

func ssoTokenNew(domain string, commenterToken string) (string, error) {
	token, err := util.RandomHex(32)
	if err != nil {
		logger.Errorf("error generating SSO token hex: %v", err)
		return "", util.ErrorInternal
	}

	statement := `
		insert into
		ssoTokens (token, domain, commenterToken, creationDate)
		values    ($1,    $2,     $3,             $4          );
	`
	_, err = svc.DB.Exec(statement, token, domain, commenterToken, time.Now().UTC())
	if err != nil {
		logger.Errorf("error inserting SSO token: %v", err)
		return "", util.ErrorInternal
	}

	return token, nil
}

// validateAuthSessionState verifies the session token initially submitted, if any, is matching the one returned with
// the given callback request
func validateAuthSessionState(sess goth.Session, req *http.Request) error {
	// Fetch the original session's URL
	rawAuthURL, err := sess.GetAuthURL()
	if err != nil {
		return err
	}

	// Parse it
	authURL, err := url.Parse(rawAuthURL)
	if err != nil {
		return err
	}

	// If there was a state initially, the value returned with the request must be the same
	if originalState := authURL.Query().Get("state"); originalState != "" {
		if reqState := req.URL.Query().Get("state"); reqState != originalState {
			logger.Debugf("Auth session state mismatch: want '%s', got '%s'", originalState, reqState)
			return errors.New("auth session state mismatch")
		}
	}
	return nil
}
