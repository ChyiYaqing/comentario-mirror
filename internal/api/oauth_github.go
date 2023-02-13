package api

import (
	"gitlab.com/comentario/comentario/internal/util"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"os"
)

var githubConfig *oauth2.Config

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
