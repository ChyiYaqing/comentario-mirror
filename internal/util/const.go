package util

// Various constants and constant-like vars

const (
	ApplicationName = "Comentario"     // Application name
	APIPath         = "api/"           // Root path of the API requests
	SwaggerUIPath   = APIPath + "docs" // Root path of the Swagger UI

	DBMaxAttempts = 10 // Max number of attempts to connect to the database
)

var (
	// UIHTMLPaths stores a list of known UI paths, which map to HTML files
	UIHTMLPaths = map[string]bool{
		"/confirm-email": true,
		"/dashboard":     true,
		"/footer":        true,
		"/forgot":        true,
		"/login":         true,
		"/logout":        true,
		"/profile":       true,
		"/reset":         true,
		"/settings":      true,
		"/signup":        true,
		"/unsubscribe":   true,
	}

	// UIPlaceholderPaths stores a list of known JS/CSS files
	UIPlaceholderPaths = map[string]bool{
		// JS
		"/js/chartist.js":    true,
		"/js/comentario.js":  true,
		"/js/count.js":       true,
		"/js/dashboard.js":   true,
		"/js/forgot.js":      true,
		"/js/highlight.js":   true,
		"/js/jquery.js":      true,
		"/js/login.js":       true,
		"/js/logout.js":      true,
		"/js/profile.js":     true,
		"/js/reset.js":       true,
		"/js/settings.js":    true,
		"/js/signup.js":      true,
		"/js/unsubscribe.js": true,
		"/js/vue.js":         true,

		// CSS
		"/css/auth.css":             true,
		"/css/auth-main.css":        true,
		"/css/button.css":           true,
		"/css/chartist.css":         true,
		"/css/checkbox.css":         true,
		"/css/comentario.css":       true,
		"/css/common-main.css":      true,
		"/css/dashboard.css":        true,
		"/css/dashboard-main.css":   true,
		"/css/email-main.css":       true,
		"/css/navbar-main.css":      true,
		"/css/source-sans.css":      true,
		"/css/tomorrow.css":         true,
		"/css/unsubscribe.css":      true,
		"/css/unsubscribe-main.css": true,
	}

	// UIStaticPaths stores a list of known UI non-HTML static paths
	UIStaticPaths = map[string]bool{
		// Fonts
		"/fonts/source-sans-300-cyrillic.woff2":     true,
		"/fonts/source-sans-300-cyrillic-ext.woff2": true,
		"/fonts/source-sans-300-greek.woff2":        true,
		"/fonts/source-sans-300-greek-ext.woff2":    true,
		"/fonts/source-sans-300-latin.woff2":        true,
		"/fonts/source-sans-300-latin-ext.woff2":    true,
		"/fonts/source-sans-300-vietnamese.woff2":   true,
		"/fonts/source-sans-400-cyrillic.woff2":     true,
		"/fonts/source-sans-400-cyrillic-ext.woff2": true,
		"/fonts/source-sans-400-greek.woff2":        true,
		"/fonts/source-sans-400-greek-ext.woff2":    true,
		"/fonts/source-sans-400-latin.woff2":        true,
		"/fonts/source-sans-400-latin-ext.woff2":    true,
		"/fonts/source-sans-400-vietnamese.woff2":   true,
		"/fonts/source-sans-700-cyrillic.woff2":     true,
		"/fonts/source-sans-700-cyrillic-ext.woff2": true,
		"/fonts/source-sans-700-greek.woff2":        true,
		"/fonts/source-sans-700-greek-ext.woff2":    true,
		"/fonts/source-sans-700-latin.woff2":        true,
		"/fonts/source-sans-700-latin-ext.woff2":    true,
		"/fonts/source-sans-700-vietnamese.woff2":   true,

		// Images
		"/images/banner.png":  true,
		"/images/favicon.ico": true,
		"/images/logo.svg":    true,
		"/images/tree.svg":    true,
	}
)
