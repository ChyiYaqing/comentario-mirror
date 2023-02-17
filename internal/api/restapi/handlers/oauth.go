package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"io"
	"net/http"
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

func OauthGithubCallback(params operations.OauthGithubCallbackParams) middleware.Responder {
	commenterToken := models.CommenterHexID(params.State)

	_, err := commenterGetByCommenterToken(commenterToken)
	if err != nil && err != util.ErrorNoSuchToken {
		return oauthFailure(err)
	}

	token, err := config.OAuthGithubConfig.Exchange(context.TODO(), params.Code)
	if err != nil {
		return oauthFailure(err)
	}

	email, err := githubGetPrimaryEmail(token.AccessToken)
	if err != nil {
		return oauthFailure(err)
	}

	resp, err := http.Get("https://api.github.com/user?access_token=" + token.AccessToken)
	if err != nil {
		return oauthFailure(err)
	}
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return oauthFailure(util.ErrorCannotReadResponse)
	}

	user := make(map[string]interface{})
	if err := json.Unmarshal(contents, &user); err != nil {
		return oauthFailure(util.ErrorInternal)
	}

	if email == "" {
		if user["email"] == nil {
			return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: "No email address returned by GitHub"})
		}
		email = user["email"].(string)
	}

	name := user["login"].(string)
	if user["name"] != nil {
		name = user["name"].(string)
	}

	link := "undefined"
	if user["html_url"] != nil {
		link = user["html_url"].(string)
	}

	photo := "undefined"
	if user["avatar_url"] != nil {
		photo = user["avatar_url"].(string)
	}

	c, err := commenterGetByEmail("github", strfmt.Email(email))
	if err != nil && err != util.ErrorNoSuchCommenter {
		return oauthFailure(err)
	}

	var commenterHex models.CommenterHexID
	if err == util.ErrorNoSuchCommenter {
		commenterHex, err = commenterNew(strfmt.Email(email), name, link, photo, "github", "")
		if err != nil {
			return oauthFailure(err)
		}
	} else {
		if err = commenterUpdate(c.CommenterHex, strfmt.Email(email), name, link, photo, "github"); err != nil {
			logger.Warningf("cannot update commenter: %s", err)
			// not a serious enough to exit with an error
		}
		commenterHex = c.CommenterHex
	}

	if err := commenterSessionUpdate(models.HexID(commenterToken), commenterHex); err != nil {
		return oauthFailure(err)
	}

	// Succeeded: close the parent window
	return closeParentWindowResponse()
}

func OauthGithubRedirect(params operations.OauthGithubRedirectParams) middleware.Responder {
	if config.OAuthGithubConfig == nil {
		return oauthNotConfigured()
	}

	_, err := commenterGetByCommenterToken(models.CommenterHexID(params.CommenterToken))
	if err != nil && err != util.ErrorNoSuchToken {
		return oauthFailure(err)
	}

	// Succeeded
	return operations.NewOauthGithubRedirectFound().WithLocation(config.OAuthGithubConfig.AuthCodeURL(params.CommenterToken))
}

func OauthGitlabCallback(params operations.OauthGitlabCallbackParams) middleware.Responder {
	commenterToken := models.CommenterHexID(params.State)

	_, err := commenterGetByCommenterToken(commenterToken)
	if err != nil && err != util.ErrorNoSuchToken {
		return oauthFailure(err)
	}

	token, err := config.OAuthGitlabConfig.Exchange(context.TODO(), params.Code)
	if err != nil {
		return oauthFailure(err)
	}

	resp, err := http.Get(config.CLIFlags.GitLabURL + "/api/v4/user?access_token=" + token.AccessToken)
	if err != nil {
		return oauthFailure(err)
	}
	logger.Infof("%v", resp.StatusCode)
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return oauthFailure(util.ErrorCannotReadResponse)
	}

	user := make(map[string]interface{})
	if err := json.Unmarshal(contents, &user); err != nil {
		return oauthFailure(util.ErrorInternal)
	}

	if user["email"] == nil {
		return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: "No email address returned by GitLab"})
	}

	email := user["email"].(string)

	if user["name"] == nil {
		return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: "No name returned by GitLab"})
	}

	name := user["name"].(string)

	link := "undefined"
	if user["web_url"] != nil {
		link = user["web_url"].(string)
	}

	photo := "undefined"
	if user["avatar_url"] != nil {
		photo = user["avatar_url"].(string)
	}

	c, err := commenterGetByEmail("gitlab", strfmt.Email(email))
	if err != nil && err != util.ErrorNoSuchCommenter {
		return oauthFailure(err)
	}

	var commenterHex models.CommenterHexID

	if err == util.ErrorNoSuchCommenter {
		commenterHex, err = commenterNew(strfmt.Email(email), name, link, photo, "gitlab", "")
		if err != nil {
			return oauthFailure(err)
		}
	} else {
		if err = commenterUpdate(c.CommenterHex, strfmt.Email(email), name, link, photo, "gitlab"); err != nil {
			logger.Warningf("cannot update commenter: %s", err)
			// not a serious enough to exit with an error
		}

		commenterHex = c.CommenterHex
	}

	if err := commenterSessionUpdate(models.HexID(commenterToken), commenterHex); err != nil {
		return oauthFailure(err)
	}

	// Succeeded: close the parent window
	return closeParentWindowResponse()
}

func OauthGitlabRedirect(params operations.OauthGitlabRedirectParams) middleware.Responder {
	if config.OAuthGitlabConfig == nil {
		return oauthNotConfigured()
	}

	_, err := commenterGetByCommenterToken(models.CommenterHexID(params.CommenterToken))
	if err != nil && err != util.ErrorNoSuchToken {
		return oauthFailure(err)
	}

	// Succeeded
	return operations.NewOauthGitlabRedirectFound().WithLocation(config.OAuthGitlabConfig.AuthCodeURL(params.CommenterToken))
}

func OauthGoogleCallback(params operations.OauthGoogleCallbackParams) middleware.Responder {
	commenterToken := models.CommenterHexID(params.State)

	_, err := commenterGetByCommenterToken(commenterToken)
	if err != nil && err != util.ErrorNoSuchToken {
		return oauthFailure(err)
	}

	token, err := config.OAuthGoogleConfig.Exchange(context.TODO(), params.Code)
	if err != nil {
		return oauthFailure(err)
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return oauthFailure(util.ErrorCannotReadResponse)
	}

	user := make(map[string]interface{})
	if err := json.Unmarshal(contents, &user); err != nil {
		return oauthFailure(util.ErrorInternal)
	}

	if user["email"] == nil {
		return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: "No email address returned by Google"})
	}

	email := user["email"].(string)

	c, err := commenterGetByEmail("google", strfmt.Email(email))
	if err != nil && err != util.ErrorNoSuchCommenter {
		return oauthFailure(err)
	}

	name := user["name"].(string)

	link := "undefined"
	if user["link"] != nil {
		link = user["link"].(string)
	}

	photo := "undefined"
	if user["picture"] != nil {
		photo = user["picture"].(string)
	}

	var commenterHex models.CommenterHexID

	if err == util.ErrorNoSuchCommenter {
		commenterHex, err = commenterNew(strfmt.Email(email), name, link, photo, "google", "")
		if err != nil {
			return oauthFailure(err)
		}
	} else {
		if err = commenterUpdate(c.CommenterHex, strfmt.Email(email), name, link, photo, "google"); err != nil {
			logger.Warningf("cannot update commenter: %s", err)
			// not a serious enough to exit with an error
		}

		commenterHex = c.CommenterHex
	}

	if err := commenterSessionUpdate(models.HexID(commenterToken), commenterHex); err != nil {
		return oauthFailure(err)
	}

	// Succeeded: close the parent window
	return closeParentWindowResponse()
}

func OauthGoogleRedirect(params operations.OauthGoogleRedirectParams) middleware.Responder {
	if config.OAuthGoogleConfig == nil {
		return oauthNotConfigured()
	}

	_, err := commenterGetByCommenterToken(models.CommenterHexID(params.CommenterToken))
	if err != nil && err != util.ErrorNoSuchToken {
		return oauthFailure(err)
	}

	// Succeeded
	return operations.NewOauthGoogleRedirectFound().WithLocation(config.OAuthGoogleConfig.AuthCodeURL(params.CommenterToken))
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
	if !d.SsoProvider {
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
	return operations.NewOauthSsoRedirectFound().WithLocation(ssoURL.String())
}

func githubGetPrimaryEmail(accessToken string) (string, error) {
	resp, err := http.Get("https://api.github.com/user/emails?access_token=" + accessToken)
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", util.ErrorCannotReadResponse
	}

	var user []map[string]interface{}
	if err := json.Unmarshal(contents, &user); err != nil {
		logger.Errorf("error unmarshalling github user: %v", err)
		return "", util.ErrorInternal
	}

	nonPrimaryEmail := ""
	for _, email := range user {
		nonPrimaryEmail = email["email"].(string)
		if email["primary"].(bool) {
			return email["email"].(string), nil
		}
	}

	return nonPrimaryEmail, nil
}

// oauthFailure returns a generic "Unauthorized" responder, with the error message in the details
func oauthFailure(err error) middleware.Responder {
	return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: err.Error()})
}

// oauthNotConfigured returns a generic "Bad request" responder, with the not-configured error message in the details
func oauthNotConfigured() middleware.Responder {
	return operations.NewGenericBadRequest().WithPayload(&operations.GenericBadRequestBody{Details: util.ErrorOAuthNotConfigured.Error()})
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
