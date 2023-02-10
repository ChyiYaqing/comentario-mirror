package api

import (
	"gitlab.com/commento/commento/api/internal/util"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"os"
)

var googleConfig *oauth2.Config

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
