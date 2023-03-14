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
var oauthSessions = &util.SafeStringMap{}

// commenterTokens maps temporary OAuth token to the related CommenterToken. It's required for those nasty identity
// providers that don't support the state parameter (such as Twitter)
var commenterTokens = &util.SafeStringMap{}

// OauthInit initiates a federated authentication process
func OauthInit(params operations.OauthInitParams) middleware.Responder {
	// Map the provider to a goth provider
	gothIdP := util.FederatedIdProviders[params.Provider]
	if gothIdP == "" {
		return operations.NewGenericBadRequest().
			WithPayload(&operations.GenericBadRequestBody{Details: "unknown provider: " + params.Provider})
	}

	// Get the registered provider instance by its name (coming from the path parameter)
	provider, err := goth.GetProvider(gothIdP)
	if err != nil {
		return operations.NewGenericBadRequest().WithPayload(&operations.GenericBadRequestBody{
			Details: fmt.Sprintf("%s (%s)", util.ErrorOAuthNotConfigured.Error(), params.Provider),
		})
	}

	// Verify the provided commenter token
	if _, err = svc.TheUserService.FindCommenterByToken(models.CommenterHexID(params.CommenterToken)); err != nil && err != svc.ErrNotFound {
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
	oauthSessions.Put(sessID, sess.Marshal())

	// If the session doesn't have the state param, also store the commenter token locally, for subsequent use
	if originalState, err := getSessionState(sess); err != nil {
		logger.Warningf("OauthInit(): failed to extract session state: %v", err)
		return operations.NewGenericInternalServerError()
	} else if originalState == "" {
		commenterTokens.Put(sessID, params.CommenterToken)
	}

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
	// Map the provider to a goth provider
	gothIdP := util.FederatedIdProviders[params.Provider]
	if gothIdP == "" {
		return operations.NewGenericBadRequest().
			WithPayload(&operations.GenericBadRequestBody{Details: "unknown provider: " + params.Provider})
	}

	// Get the registered provider instance by its name (coming from the path parameter)
	provider, err := goth.GetProvider(gothIdP)
	if err != nil {
		logger.Debugf("Failed to fetch provider '%s': %v", params.Provider, err)
		return oauthFailure(fmt.Errorf("provider not configured: %s", params.Provider))
	}

	// Obtain the auth session ID from the cookie
	var sess goth.Session
	var sessID string
	if cookie, err := params.HTTPRequest.Cookie(util.AuthSessionCookieName); err != nil {
		logger.Debugf("Auth session cookie error: %v", err)
		return oauthFailure(errors.New("auth session cookie missing"))
	} else {
		sessID = cookie.Value
	}

	// Find and delete the session
	if sessData, ok := oauthSessions.Take(sessID); !ok {
		logger.Debugf("No auth session found with ID=%v: %v", sessID, err)
		return oauthFailure(errors.New("auth session not found"))

		// Recover the original provider session
	} else if sess, err = provider.UnmarshalSession(sessData); err != nil {
		logger.Debugf("provider.UnmarshalSession() failed: %v", err)
		return oauthFailure(errors.New("auth session unmarshalling"))
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

	// Obtain the commenter token: if it isn't present in the state param (Twitter doesn't support state), try to find
	// it in the token store
	commenterToken := reqParams.Get("state")
	if commenterToken == "" {
		commenterToken, _ = commenterTokens.Take(sessID)
	}
	if commenterToken == "" {
		return oauthFailure(errors.New("failed to obtain commenter token"))
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
	// -- Avatar
	avatar := fedUser.AvatarURL
	if avatar == "" {
		// TODO get rid of this crap
		avatar = "undefined"
	}

	// Try to find the corresponding commenter by their token
	if _, err := svc.TheUserService.FindCommenterByToken(models.CommenterHexID(commenterToken)); err != nil && err != svc.ErrNotFound {
		return oauthFailure(err)
	}

	// Now try to find an existing commenter by their email
	var commenterHex models.CommenterHexID
	if commenter, err := svc.TheUserService.FindCommenterByIdPEmail(params.Provider, fedUser.Email); err == nil {
		// Commenter found
		commenterHex = commenter.CommenterHexID()

	} else if err != svc.ErrNotFound {
		// Any other error than "not found"
		return oauthFailure(err)
	}

	// No such commenter yet: it's a signup
	if commenterHex == "" {
		// Create a new commenter
		if commenterHex, err = commenterNew(strfmt.Email(fedUser.Email), fedUser.Name, "undefined", avatar, params.Provider, ""); err != nil {
			return oauthFailure(err)
		}

		// Commenter already exists: it's a login. Update commenter's details
	} else if err = commenterUpdate(commenterHex, strfmt.Email(fedUser.Email), fedUser.Name, "undefined", avatar, params.Provider); err != nil {
		// Failed to update, but proceed nonetheless
		logger.Warningf("cannot update commenter: %s", err)
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

	// Try to find the corresponding commenter by their token
	if _, err := svc.TheUserService.FindCommenterByToken(commenterToken); err != nil && err != svc.ErrNotFound {
		return oauthFailure(err)
	}

	// Now try to find an existing commenter by their email
	var commenterHex models.CommenterHexID
	idp := "sso:" + domain
	if commenter, err := svc.TheUserService.FindCommenterByIdPEmail(idp, payload.Email); err == nil {
		// Commenter found
		commenterHex = commenter.CommenterHexID()

	} else if err != svc.ErrNotFound {
		// Any other error than "not found"
		return oauthFailure(err)
	}

	// No such commenter yet: it's a signup
	if commenterHex == "" {
		if commenterHex, err = commenterNew(strfmt.Email(payload.Email), payload.Name, payload.Link, payload.Photo, idp, ""); err != nil {
			return oauthFailure(err)
		}
	} else if err = commenterUpdate(commenterHex, strfmt.Email(payload.Email), payload.Name, payload.Link, payload.Photo, idp); err != nil {
		// Failed to update, but proceed nonetheless
		logger.Warningf("cannot update commenter: %s", err)
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

	if _, err = svc.TheUserService.FindCommenterByToken(models.CommenterHexID(params.CommenterToken)); err != nil && err != svc.ErrNotFound {
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

// getSessionState extracts the state parameter from the given session's URL
func getSessionState(sess goth.Session) (string, error) {
	// Fetch the original session's URL
	rawAuthURL, err := sess.GetAuthURL()
	if err != nil {
		return "", err
	}

	// Parse it
	authURL, err := url.Parse(rawAuthURL)
	if err != nil {
		return "", err
	}

	// Extract the state param
	return authURL.Query().Get("state"), nil
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
	row := svc.DB.QueryRow("select domain, commenterToken from ssoTokens where token = $1;", token)

	var domain string
	var commenterToken models.CommenterHexID
	if err := row.Scan(&domain, &commenterToken); err != nil {
		return "", "", util.ErrorNoSuchToken
	}

	if err := svc.DB.Exec("delete from ssoTokens where token = $1;", token); err != nil {
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

	err = svc.DB.Exec(
		"insert into ssoTokens(token, domain, commenterToken, creationDate) values($1, $2, $3, $4);",
		token,
		domain,
		commenterToken,
		time.Now().UTC())
	if err != nil {
		logger.Errorf("error inserting SSO token: %v", err)
		return "", util.ErrorInternal
	}

	return token, nil
}

// validateAuthSessionState verifies the session token initially submitted, if any, is matching the one returned with
// the given callback request
func validateAuthSessionState(sess goth.Session, req *http.Request) error {
	// Extract the original session state
	originalState, err := getSessionState(sess)
	if err != nil {
		return err
	}

	// If there was a state initially, the value returned with the request must be the same
	if originalState != "" {
		if reqState := req.URL.Query().Get("state"); reqState != originalState {
			logger.Debugf("Auth session state mismatch: want '%s', got '%s'", originalState, reqState)
			return errors.New("auth session state mismatch")
		}
	}
	return nil
}
