package util

import "time"

// Various constants and constant-like vars

const (
	ApplicationName = "Comentario"     // Application name
	APIPath         = "api/"           // Root path of the API requests
	SwaggerUIPath   = APIPath + "docs" // Root path of the Swagger UI

	DBMaxAttempts = 10 // Max number of attempts to connect to the database

	AuthSessionCookieName = "_comentario-auth-session" // Cookie name to store the federated authentication session ID
)

// StaticEntry holds the information about a static resource served over HTTP
type StaticEntry struct {
	Src  string // Source file pattern, using '$' for the base name (last path element). If empty, the original path is used
	Repl bool   // Whether to replace placeholders ('[[[.xxx]]]') in the file
}

var (
	WrongAuthDelay = 10 * time.Second // Delay to exercise on a wrong email, password etc.

	// FederatedIdProviders maps all known federated identity providers from our IDs to goth IDs
	FederatedIdProviders = map[string]string{
		"github":  "github",
		"gitlab":  "gitlab",
		"google":  "google",
		"twitter": "twitter",
	}

	// UIStaticPaths stores a map of known UI static paths to their info entries
	UIStaticPaths = map[string]StaticEntry{
		// Root
		"favicon.ico": {Src: "images/$"},

		// Fonts
		"fonts/source-sans-300-cyrillic.woff2":     {},
		"fonts/source-sans-300-cyrillic-ext.woff2": {},
		"fonts/source-sans-300-greek.woff2":        {},
		"fonts/source-sans-300-greek-ext.woff2":    {},
		"fonts/source-sans-300-latin.woff2":        {},
		"fonts/source-sans-300-latin-ext.woff2":    {},
		"fonts/source-sans-300-vietnamese.woff2":   {},
		"fonts/source-sans-400-cyrillic.woff2":     {},
		"fonts/source-sans-400-cyrillic-ext.woff2": {},
		"fonts/source-sans-400-greek.woff2":        {},
		"fonts/source-sans-400-greek-ext.woff2":    {},
		"fonts/source-sans-400-latin.woff2":        {},
		"fonts/source-sans-400-latin-ext.woff2":    {},
		"fonts/source-sans-400-vietnamese.woff2":   {},
		"fonts/source-sans-700-cyrillic.woff2":     {},
		"fonts/source-sans-700-cyrillic-ext.woff2": {},
		"fonts/source-sans-700-greek.woff2":        {},
		"fonts/source-sans-700-greek-ext.woff2":    {},
		"fonts/source-sans-700-latin.woff2":        {},
		"fonts/source-sans-700-latin-ext.woff2":    {},
		"fonts/source-sans-700-vietnamese.woff2":   {},

		// Images
		"images/banner.png": {},
		"images/logo.svg":   {},
		"images/tree.svg":   {},

		// HTML
		"confirm-email": {Src: "html/$.html", Repl: true},
		"dashboard":     {Src: "html/$.html", Repl: true},
		"footer":        {Src: "html/$.html", Repl: true},
		"forgot":        {Src: "html/$.html", Repl: true},
		"login":         {Src: "html/$.html", Repl: true},
		"logout":        {Src: "html/$.html", Repl: true},
		"profile":       {Src: "html/$.html", Repl: true},
		"reset":         {Src: "html/$.html", Repl: true},
		"settings":      {Src: "html/$.html", Repl: true},
		"signup":        {Src: "html/$.html", Repl: true},
		"unsubscribe":   {Src: "html/$.html", Repl: true},

		// JS
		"js/chartist.js":    {Repl: true},
		"js/comentario.js":  {Repl: true},
		"js/count.js":       {Repl: true},
		"js/dashboard.js":   {Repl: true},
		"js/forgot.js":      {Repl: true},
		"js/highlight.js":   {Repl: true},
		"js/jquery.js":      {Repl: true},
		"js/login.js":       {Repl: true},
		"js/logout.js":      {Repl: true},
		"js/profile.js":     {Repl: true},
		"js/reset.js":       {Repl: true},
		"js/settings.js":    {Repl: true},
		"js/signup.js":      {Repl: true},
		"js/unsubscribe.js": {Repl: true},
		"js/vue.js":         {Repl: true},

		// CSS
		"css/auth.css":             {Repl: true},
		"css/auth-main.css":        {Repl: true},
		"css/button.css":           {Repl: true},
		"css/chartist.css":         {Repl: true},
		"css/checkbox.css":         {Repl: true},
		"css/comentario.css":       {Repl: true},
		"css/common-main.css":      {Repl: true},
		"css/dashboard.css":        {Repl: true},
		"css/dashboard-main.css":   {Repl: true},
		"css/email-main.css":       {Repl: true},
		"css/navbar-main.css":      {Repl: true},
		"css/source-sans.css":      {Repl: true},
		"css/tomorrow.css":         {Repl: true},
		"css/unsubscribe.css":      {Repl: true},
		"css/unsubscribe-main.css": {Repl: true},
	}
)
