package util

import (
	"golang.org/x/text/language"
	"time"
)

// Various constants and constant-like vars

const (
	ApplicationName = "Comentario"     // Application name
	APIPath         = "api/"           // Root path of the API requests
	SwaggerUIPath   = APIPath + "docs" // Root path of the Swagger UI

	OneDay = 24 * time.Hour // Time unit representing one day

	DBMaxAttempts = 10 // Max number of attempts to connect to the database

	CookieNameUserToken   = "comentario_user_token"    // Cookie name to store the token of the authenticated (owner) user
	CookieNameAuthSession = "_comentario_auth_session" // Cookie name to store the federated authentication session ID
	LangCookieDuration    = 365 * OneDay               // How long the language cookie stays valid
	HeaderCommenterToken  = "X-Commenter-Token"        // Name of the header that contains the token of the authenticated commenter user
)

var (
	WrongAuthDelay = 10 * time.Second // Delay to exercise on a wrong email, password etc.

	// FederatedIdProviders maps all known federated identity providers from our IDs to goth IDs
	FederatedIdProviders = map[string]string{
		"github":  "github",
		"gitlab":  "gitlab",
		"google":  "google",
		"twitter": "twitter",
	}

	// UILanguageTags stores tags of supported frontend languages
	UILanguageTags = []language.Tag{
		language.English, // The first language is used as fallback
	}

	// UILangMatcher is a matcher instance for UI languages
	UILangMatcher = language.NewMatcher(UILanguageTags)

	// UIStaticPaths stores a map of known UI static paths to a flag that says whether the file contains replacements
	UIStaticPaths = map[string]bool{
		"favicon.ico":    false,
		"comentario.js":  true,
		"comentario.css": true,
	}
)
