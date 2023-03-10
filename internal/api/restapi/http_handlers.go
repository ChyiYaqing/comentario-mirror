package restapi

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
)

// InternalError responds with the "Internal server error" response
func InternalError(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// notFoundBypassWriter is an object that pretends to be a ResponseWriter but refrains from writing a 404 response
type notFoundBypassWriter struct {
	http.ResponseWriter
	status int
}

func (w *notFoundBypassWriter) WriteHeader(status int) {
	// Store the status for our own use
	w.status = status

	// Pass through unless it's a NotFound response
	if status != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *notFoundBypassWriter) Write(p []byte) (int, error) {
	// Do not write anything on a NotFound response, but pretend the write has been successful
	if w.status == http.StatusNotFound {
		return len(p), nil
	}

	// Pass through to the real writer
	return w.ResponseWriter.Write(p)
}

// corsHandler returns a middleware that adds CORS headers to responses
func corsHandler(next http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedHeaders([]string{"Content-Type", "X-Requested-With"}),
		handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost}))(next)
}

// fallbackHandler returns a middleware that is called in case all other handlers failed
func fallbackHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only serve 404's on GET requests
		if r.Method == http.MethodGet {
			http.NotFoundHandler().ServeHTTP(w, r)

		} else {
			// Any other method
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("Method not allowed"))
		}
	})
}

// makeAPIHandler returns a constructor function for the provided API handler
func makeAPIHandler(apiHandler http.Handler) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify the URL is a correct one
			if ok, p := config.PathOfBaseURL(r.URL.Path); ok {
				// If it's an API call. Also check whether the Swagger UI is enabled (because it's also served by the API)
				isAPIPath := strings.HasPrefix(p, util.APIPath)
				isSwaggerPath := p == "swagger.json" || strings.HasPrefix(p, util.SwaggerUIPath)
				if !isSwaggerPath && isAPIPath || isSwaggerPath && config.CLIFlags.EnableSwaggerUI {
					r.URL.Path = "/" + p
					apiHandler.ServeHTTP(w, r)
					return
				}
			}

			// Pass on to the next handler otherwise
			next.ServeHTTP(w, r)
		})
	}
}

// redirectToLangRootHandler returns a middleware that redirects the user from the site root or an "incomplete" language
// root (such as "/en") to the complete/appropriate language root (such as "/en/")
func redirectToLangRootHandler(next http.Handler) http.Handler {
	// Replace the path in the provided URL and return the whole URL as a string
	replacePath := func(u *url.URL, p string) string {
		// Clone the URL
		cu := *u
		// Wipe out any user info
		cu.User = nil
		// Replace the path
		cu.Path = p
		return cu.String()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If it's 'GET <path-under-root>'
		if ok, p := config.PathOfBaseURL(r.URL.Path); ok && r.Method == http.MethodGet {
			switch len(p) {
			// Site root: redirect to the most appropriate language root
			case 0:
				// Redirect with 302 and not "Moved Permanently" to avoid caching by browsers
				http.Redirect(
					w,
					r,
					replacePath(r.URL, fmt.Sprintf("/%s/", config.GuessUserLanguage(r))),
					http.StatusFound)
				return

			// If it's an "incomplete" language root, redirect to the full root, permanently
			case 2:
				if util.IsUILang(p) {
					http.Redirect(
						w,
						r,
						replacePath(r.URL, fmt.Sprintf("/%s/", p)), http.StatusMovedPermanently)
					return
				}
			}
		}

		// Otherwise, hand over to the next handler
		next.ServeHTTP(w, r)
	})
}

// serveFileWithPlaceholders serves out files that contain placeholders, ie. HTML, CSS, and JS files
func serveFileWithPlaceholders(filePath string, w http.ResponseWriter) {
	logger.Debugf("Serving file /%s replacing placeholders", filePath)

	// Read in the file
	filename := path.Join(config.CLIFlags.StaticPath, filePath)
	b, err := os.ReadFile(filename)
	if err != nil {
		logger.Warningf("Failed to read %s: %v", filename, err)
		InternalError(w)
		return
	}

	// Pass the file through the replacements, if there's a placeholder found
	s := string(b)
	if strings.Contains(s, "[[[.") {
		b = []byte(
			strings.Replace(strings.Replace(strings.Replace(strings.Replace(s,
				"[[[.Origin]]]", strings.TrimSuffix(config.BaseURL.String(), "/"), -1),
				"[[[.CdnPrefix]]]", strings.TrimSuffix(config.CDNURL.String(), "/"), -1),
				"[[[.Footer]]]", "TODO footer", -1),
				"[[[.Version]]]", config.AppVersion, -1))
	}

	// Determine content type
	ctype := "text/plain"
	if strings.HasSuffix(filePath, ".html") {
		ctype = "text/html; charset=utf-8"
	} else if strings.HasSuffix(filePath, ".js") {
		ctype = "text/javascript; charset=utf-8"
	} else if strings.HasSuffix(filePath, ".css") {
		ctype = "text/css; charset=utf-8"
	}

	// Serve the final result out
	h := w.Header()
	h.Set("Content-Length", strconv.Itoa(len(b)))
	h.Set("Content-Type", ctype)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

// staticHandler returns a middleware that serves the static content of the app, which includes:
// - stuff listed in UIStaticPaths[] (favicon and such)
// - paths starting from a language root ('/en/', '/ru/' etc.)
func staticHandler(next http.Handler) http.Handler {
	// Instantiate a file server for static content
	fileHandler := http.FileServer(http.Dir(config.CLIFlags.StaticPath))

	// Make a middleware handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Resources are only served out via GET
		if r.Method == http.MethodGet {
			// If it's a path under the base URL
			if ok, p := config.PathOfBaseURL(r.URL.Path); ok {
				// Check if it's a static resource or a path on/under a language root
				repl, static := util.UIStaticPaths[p]
				hasLang := !static && len(p) >= 3 && p[2] == '/' && util.IsUILang(p[0:2])
				langRoot := hasLang && len(p) == 3 // If the path looks like 'xx/', it's a language root

				// If under a language root, set a language cookie
				if hasLang {
					http.SetCookie(w, &http.Cookie{
						Name:   "lang",
						Value:  p[0:2],
						Path:   "/",
						MaxAge: int(util.LangCookieDuration.Seconds()),
					})
				}

				// If it's a static file with placeholders, serve it with replacements
				if static && repl {
					serveFileWithPlaceholders(p, w)
					return
				}

				// Non-replaceable static stuff and content under language root
				if static || hasLang {
					// Do not allow directory browsing (any path ending with a '/', which isn't a language root)
					if !langRoot && !strings.HasSuffix(p, "/") {
						// Make a "fake" (bypassing) response writer
						bypassWriter := &notFoundBypassWriter{ResponseWriter: w}

						// Try to serve the requested static content
						fileHandler.ServeHTTP(bypassWriter, r)

						// If the content was found, we're done
						if bypassWriter.status != http.StatusNotFound {
							return
						}
					}

					// Language root or file wasn't found: serve the main application script for the given language
					if hasLang {
						serveFileWithPlaceholders(fmt.Sprintf("%s/index.html", p[0:2]), w)
						return
					}

					// Remove any existing content type to allow automatic MIME type detection
					delete(w.Header(), "Content-Type")
				}
			}
		}

		// Pass on to the next handler otherwise
		next.ServeHTTP(w, r)
	})
}
