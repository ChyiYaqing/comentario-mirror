package handlers

import (
	"gitlab.com/comentario/comentario/internal/util"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/gitlab"
	"golang.org/x/oauth2/google"
	"os"
)

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

var githubConfig *oauth2.Config
var gitlabConfig *oauth2.Config
var googleConfig *oauth2.Config

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
