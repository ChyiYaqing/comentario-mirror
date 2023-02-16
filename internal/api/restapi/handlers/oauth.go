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
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/gitlab"
	"golang.org/x/oauth2/google"
	"io"
	"net/http"
	"net/url"
	"os"
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

var googleConfigured bool
var githubConfigured bool
var gitlabConfigured bool

func OAuthConfigure() error {
	if err := googleOauthConfigure(); err != nil {
		return err
	}
	if err := githubOauthConfigure(); err != nil {
		return err
	}
	if err := gitlabOauthConfigure(); err != nil {
		return err
	}
	return nil
}

func OauthGithubCallback(params operations.OauthGithubCallbackParams) middleware.Responder {
	commenterToken := models.CommenterToken(params.State)

	_, err := commenterGetByCommenterToken(commenterToken)
	if err != nil && err != util.ErrorNoSuchToken {
		return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: err.Error()})
	}

	token, err := githubConfig.Exchange(context.TODO(), params.Code)
	if err != nil {
		return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: err.Error()})
	}

	email, err := githubGetPrimaryEmail(token.AccessToken)
	if err != nil {
		return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: err.Error()})
	}

	resp, err := http.Get("https://api.github.com/user?access_token=" + token.AccessToken)
	if err != nil {
		return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: err.Error()})
	}
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: util.ErrorCannotReadResponse.Error()})
	}

	user := make(map[string]interface{})
	if err := json.Unmarshal(contents, &user); err != nil {
		return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: util.ErrorInternal.Error()})
	}

	if email == "" {
		if user["email"] == nil {
			return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: "Error: no email address returned by Github"})
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
		return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: err.Error()})
	}

	var commenterHex models.HexID
	if err == util.ErrorNoSuchCommenter {
		commenterHex, err = commenterNew(strfmt.Email(email), name, link, photo, "github", "")
		if err != nil {
			return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: err.Error()})
		}
	} else {
		if err = commenterUpdate(c.CommenterHex, strfmt.Email(email), name, link, photo, "github"); err != nil {
			logger.Warningf("cannot update commenter: %s", err)
			// not a serious enough to exit with an error
		}
		commenterHex = c.CommenterHex
	}

	if err := commenterSessionUpdate(models.HexID(commenterToken), commenterHex); err != nil {
		return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: err.Error()})
	}

	// Succeeded: close the parent window
	return closeParentWindowResponse()
}

func OauthGithubRedirect(params operations.OauthGithubRedirectParams) middleware.Responder {
	if githubConfig == nil {
		return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: "Authentication via GitHub is not configured"})
	}

	_, err := commenterGetByCommenterToken(models.CommenterToken(params.CommenterToken))
	if err != nil && err != util.ErrorNoSuchToken {
		return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: err.Error()})
	}

	return operations.NewOauthGithubRedirectFound().WithLocation(githubConfig.AuthCodeURL(params.CommenterToken))
}

func OauthGitlabCallback(params operations.OauthGitlabCallbackParams) middleware.Responder {
	commenterToken := models.CommenterToken(params.State)

	_, err := commenterGetByCommenterToken(commenterToken)
	if err != nil && err != util.ErrorNoSuchToken {
		_, _ = fmt.Fprintf(w, "Error: %s\n", err.Error())
		return
	}

	token, err := gitlabConfig.Exchange(context.TODO(), params.Code)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}

	resp, err := http.Get(os.Getenv("GITLAB_URL") + "/api/v4/user?access_token=" + token.AccessToken)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}
	logger.Infof("%v", resp.StatusCode)
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s", util.ErrorCannotReadResponse.Error())
		return
	}

	user := make(map[string]interface{})
	if err := json.Unmarshal(contents, &user); err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s", util.ErrorInternal.Error())
		return
	}

	if user["email"] == nil {
		_, _ = fmt.Fprintf(w, "Error: no email address returned by Gitlab")
		return
	}

	email := user["email"].(string)

	if user["name"] == nil {
		_, _ = fmt.Fprintf(w, "Error: no name returned by Gitlab")
		return
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
		_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}

	var commenterHex string

	if err == util.ErrorNoSuchCommenter {
		commenterHex, err = commenterNew(strfmt.Email(email), name, link, photo, "gitlab", "")
		if err != nil {
			_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
	} else {
		if err = commenterUpdate(c.CommenterHex, email, name, link, photo, "gitlab"); err != nil {
			logger.Warningf("cannot update commenter: %s", err)
			// not a serious enough to exit with an error
		}

		commenterHex = c.CommenterHex
	}

	if err := commenterSessionUpdate(commenterToken, commenterHex); err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}

	// Succeeded: close the parent window
	return closeParentWindowResponse()
}

func OauthGitlabRedirect(params operations.OauthGitlabRedirectParams) middleware.Responder {
	if gitlabConfig == nil {
		logger.Errorf("gitlab oauth access attempt without configuration")
		_, _ = fmt.Fprintf(w, "error: this website has not configured gitlab OAuth")
		return
	}

	commenterToken := r.FormValue("commenterToken")

	_, err := commenterGetByCommenterToken(commenterToken)
	if err != nil && err != util.ErrorNoSuchToken {
		_, _ = fmt.Fprintf(w, "error: %s\n", err.Error())
		return
	}

	return operations.NewOauthGitlabRedirectFound().WithLocation(gitlabConfig.AuthCodeURL(params.CommenterToken))
}

func OauthGoogleCallback(params operations.OauthGoogleCallbackParams) middleware.Responder {
	commenterToken := models.CommenterToken(params.State)

	_, err := commenterGetByCommenterToken(commenterToken)
	if err != nil && err != util.ErrorNoSuchToken {
		_, _ = fmt.Fprintf(w, "Error: %s\n", err.Error())
		return
	}

	token, err := googleConfig.Exchange(context.TODO(), params.Code)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s", util.ErrorCannotReadResponse.Error())
		return
	}

	user := make(map[string]interface{})
	if err := json.Unmarshal(contents, &user); err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s", util.ErrorInternal.Error())
		return
	}

	if user["email"] == nil {
		_, _ = fmt.Fprintf(w, "Error: no email address returned by Github")
		return
	}

	email := user["email"].(string)

	c, err := commenterGetByEmail("google", strfmt.Email(email))
	if err != nil && err != util.ErrorNoSuchCommenter {
		_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
		return
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

	var commenterHex string

	if err == util.ErrorNoSuchCommenter {
		commenterHex, err = commenterNew(strfmt.Email(email), name, link, photo, "google", "")
		if err != nil {
			_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
	} else {
		if err = commenterUpdate(c.CommenterHex, email, name, link, photo, "google"); err != nil {
			logger.Warningf("cannot update commenter: %s", err)
			// not a serious enough to exit with an error
		}

		commenterHex = c.CommenterHex
	}

	if err := commenterSessionUpdate(commenterToken, commenterHex); err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}

	// Succeeded: close the parent window
	return closeParentWindowResponse()
}

func OauthGoogleRedirect(params operations.OauthGoogleRedirectParams) middleware.Responder {
	if googleConfig == nil {
		logger.Errorf("google oauth access attempt without configuration")
		_, _ = fmt.Fprintf(w, "error: this website has not configured Google OAuth")
		return
	}

	commenterToken := r.FormValue("commenterToken")

	_, err := commenterGetByCommenterToken(commenterToken)
	if err != nil && err != util.ErrorNoSuchToken {
		_, _ = fmt.Fprintf(w, "error: %s\n", err.Error())
		return
	}

	return operations.NewOauthGoogleRedirectFound().WithLocation(googleConfig.AuthCodeURL(params.CommenterToken))
}

func OauthSsoCallback(params operations.OauthSsoCallbackParams) middleware.Responder {
	payloadHex := r.FormValue("payload")
	signature := r.FormValue("hmac")

	payloadBytes, err := hex.DecodeString(payloadHex)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: invalid JSON payload hex encoding: %s\n", err.Error())
		return
	}

	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: invalid HMAC signature hex encoding: %s\n", err.Error())
		return
	}

	payload := ssoPayload{}
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: cannot unmarshal JSON payload: %s\n", err.Error())
		return
	}

	if payload.Token == "" || payload.Email == "" || payload.Name == "" {
		_, _ = fmt.Fprintf(w, "Error: %s\n", util.ErrorMissingField.Error())
		return
	}

	if payload.Link == "" {
		payload.Link = "undefined"
	}

	if payload.Photo == "" {
		payload.Photo = "undefined"
	}

	domain, commenterToken, err := ssoTokenExtract(payload.Token)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s\n", err.Error())
		return
	}

	d, err := domainGet(domain)
	if err != nil {
		if err == util.ErrorNoSuchDomain {
			_, _ = fmt.Fprintf(w, "Error: %s\n", err.Error())
		} else {
			logger.Errorf("cannot get domain for SSO: %v", err)
			_, _ = fmt.Fprintf(w, "Error: %s\n", util.ErrorInternal.Error())
		}
		return
	}

	if d.SsoSecret == "" || d.SsoUrl == "" {
		_, _ = fmt.Fprintf(w, "Error: %s\n", util.ErrorMissingConfig.Error())
		return
	}

	key, err := hex.DecodeString(d.SsoSecret)
	if err != nil {
		logger.Errorf("cannot decode SSO secret as hex: %v", err)
		_, _ = fmt.Fprintf(w, "Error: %s\n", err.Error())
		return
	}

	h := hmac.New(sha256.New, key)
	h.Write(payloadBytes)
	expectedSignatureBytes := h.Sum(nil)
	if !hmac.Equal(expectedSignatureBytes, signatureBytes) {
		_, _ = fmt.Fprintf(w, "Error: HMAC signature verification failed\n")
		return
	}

	_, err = commenterGetByCommenterToken(commenterToken)
	if err != nil && err != util.ErrorNoSuchToken {
		_, _ = fmt.Fprintf(w, "Error: %s\n", err.Error())
		return
	}

	c, err := commenterGetByEmail("sso:"+domain, strfmt.Email(payload.Email))
	if err != nil && err != util.ErrorNoSuchCommenter {
		_, _ = fmt.Fprintf(w, "Error: %s\n", err.Error())
		return
	}

	var commenterHex string

	if err == util.ErrorNoSuchCommenter {
		commenterHex, err = commenterNew(strfmt.Email(payload.Email), payload.Name, payload.Link, payload.Photo, "sso:"+domain, "")
		if err != nil {
			_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
	} else {
		if err = commenterUpdate(c.CommenterHex, payload.Email, payload.Name, payload.Link, payload.Photo, "sso:"+domain); err != nil {
			logger.Warningf("cannot update commenter: %s", err)
			// not a serious enough to exit with an error
		}

		commenterHex = c.CommenterHex
	}

	if err = commenterSessionUpdate(commenterToken, commenterHex); err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s\n", err.Error())
		return
	}

	// Succeeded: close the parent window
	return closeParentWindowResponse()
}

func OauthSsoRedirect(params operations.OauthSsoRedirectParams) middleware.Responder {
	commenterToken := r.FormValue("commenterToken")
	domain := r.Header.Get("Referer")

	if commenterToken == "" {
		_, _ = fmt.Fprintf(w, "Error: %s\n", util.ErrorMissingField.Error())
		return
	}

	domain = domainStrip(domain)
	if domain == "" {
		_, _ = fmt.Fprintf(w, "Error: No Referer header found in request\n")
		return
	}

	_, err := commenterGetByCommenterToken(commenterToken)
	if err != nil && err != util.ErrorNoSuchToken {
		_, _ = fmt.Fprintf(w, "Error: %s\n", err.Error())
		return
	}

	d, err := domainGet(domain)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s\n", util.ErrorNoSuchDomain.Error())
		return
	}

	if !d.SsoProvider {
		_, _ = fmt.Fprintf(w, "Error: SSO not configured for %s\n", domain)
		return
	}

	if d.SsoSecret == "" || d.SsoUrl == "" {
		_, _ = fmt.Fprintf(w, "Error: %s\n", util.ErrorMissingConfig.Error())
		return
	}

	key, err := hex.DecodeString(d.SsoSecret)
	if err != nil {
		logger.Errorf("cannot decode SSO secret as hex: %v", err)
		_, _ = fmt.Fprintf(w, "Error: %s\n", err.Error())
		return
	}

	token, err := ssoTokenNew(domain, commenterToken)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s\n", err.Error())
		return
	}

	tokenBytes, err := hex.DecodeString(token)
	if err != nil {
		logger.Errorf("cannot decode hex token: %v", err)
		_, _ = fmt.Fprintf(w, "Error: %s\n", util.ErrorInternal.Error())
		return
	}

	h := hmac.New(sha256.New, key)
	h.Write(tokenBytes)
	signature := hex.EncodeToString(h.Sum(nil))

	u, err := url.Parse(d.SsoUrl)
	if err != nil {
		// this should really not be happening; we're checking if the
		// passed URL is valid at domain update
		logger.Errorf("cannot parse URL: %v", err)
		_, _ = fmt.Fprintf(w, "Error: %s\n", util.ErrorInternal.Error())
		return
	}

	q := u.Query()
	q.Set("token", token)
	q.Set("hmac", signature)
	u.RawQuery = q.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)
}

var githubConfig *oauth2.Config
var gitlabConfig *oauth2.Config
var googleConfig *oauth2.Config

func githubGetPrimaryEmail(accessToken string) (string, error) {
	resp, err := http.Get("https://api.github.com/user/emails?access_token=" + accessToken)
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", util.ErrorCannotReadResponse
	}

	var user []map[string]interface{}
	if err := json.Unmarshal(contents, &user); err != nil {
		logger.Errorf("error unmarshaling github user: %v", err)
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

func githubOauthConfigure() error {
	githubConfig = nil
	if os.Getenv("GITHUB_KEY") == "" && os.Getenv("GITHUB_SECRET") == "" {
		return nil
	}

	if os.Getenv("GITHUB_KEY") == "" {
		logger.Errorf("COMENTARIO_GITHUB_KEY not configured, but COMENTARIO_GITHUB_SECRET is set")
		return util.ErrorOauthMisconfigured
	}

	if os.Getenv("GITHUB_SECRET") == "" {
		logger.Errorf("COMENTARIO_GITHUB_SECRET not configured, but COMENTARIO_GITHUB_KEY is set")
		return util.ErrorOauthMisconfigured
	}

	logger.Infof("loading github OAuth config")

	githubConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("ORIGIN") + "/api/oauth/github/callback",
		ClientID:     os.Getenv("GITHUB_KEY"),
		ClientSecret: os.Getenv("GITHUB_SECRET"),
		Scopes: []string{
			"read:user",
			"user:email",
		},
		Endpoint: github.Endpoint,
	}

	githubConfigured = true
	return nil
}

func gitlabOauthConfigure() error {
	gitlabConfig = nil
	if os.Getenv("GITLAB_KEY") == "" && os.Getenv("GITLAB_SECRET") == "" {
		return nil
	}

	if os.Getenv("GITLAB_KEY") == "" {
		logger.Errorf("COMENTARIO_GITLAB_KEY not configured, but COMENTARIO_GITLAB_SECRET is set")
		return util.ErrorOauthMisconfigured
	}

	if os.Getenv("GITLAB_SECRET") == "" {
		logger.Errorf("COMENTARIO_GITLAB_SECRET not configured, but COMENTARIO_GITLAB_KEY is set")
		return util.ErrorOauthMisconfigured
	}

	logger.Infof("loading gitlab OAuth config")

	gitlabConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("ORIGIN") + "/api/oauth/gitlab/callback",
		ClientID:     os.Getenv("GITLAB_KEY"),
		ClientSecret: os.Getenv("GITLAB_SECRET"),
		Scopes: []string{
			"read_user",
		},
		Endpoint: gitlab.Endpoint,
	}
	gitlabConfig.Endpoint.AuthURL = os.Getenv("GITLAB_URL") + "/oauth/authorize"
	gitlabConfig.Endpoint.TokenURL = os.Getenv("GITLAB_URL") + "/oauth/token"

	gitlabConfigured = true
	return nil
}

func googleOauthConfigure() error {
	googleConfig = nil
	if os.Getenv("GOOGLE_KEY") == "" && os.Getenv("GOOGLE_SECRET") == "" {
		return nil
	}

	if os.Getenv("GOOGLE_KEY") == "" {
		logger.Errorf("COMENTARIO_GOOGLE_KEY not configured, but COMENTARIO_GOOGLE_SECRET is set")
		return util.ErrorOauthMisconfigured
	}

	if os.Getenv("GOOGLE_SECRET") == "" {
		logger.Errorf("COMENTARIO_GOOGLE_SECRET not configured, but COMENTARIO_GOOGLE_KEY is set")
		return util.ErrorOauthMisconfigured
	}

	logger.Infof("loading Google OAuth config")

	googleConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("ORIGIN") + "/api/oauth/google/callback",
		ClientID:     os.Getenv("GOOGLE_KEY"),
		ClientSecret: os.Getenv("GOOGLE_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

	googleConfigured = true
	return nil
}

func ssoTokenExtract(token string) (string, string, error) {
	statement := "select domain, commenterToken from ssoTokens where token = $1;"
	row := svc.DB.QueryRow(statement, token)

	var domain string
	var commenterToken string
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
