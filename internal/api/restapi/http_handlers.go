package restapi

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
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
		handlers.AllowedHeaders([]string{"X-Requested-With"}),
		handlers.AllowedMethods([]string{"GET", "POST"}))(next)
}

// fallbackHandler returns a middleware that is called in case all other handlers failed
func fallbackHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only serve 404's on GET requests
		if r.Method == "GET" {
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

// rootRedirectHandler returns a middleware that redirects to login from the root path ("/")
func rootRedirectHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If it's 'GET <root>', redirect to login
		if ok, p := config.PathOfBaseURL(r.URL.Path); ok && p == "" && r.Method == "GET" {
			logger.Debug("Redirecting to login")
			http.Redirect(w, r, config.URLFor("login", nil), 301)
			return
		}

		// Pass on to the next handler otherwise
		next.ServeHTTP(w, r)
	})
}

// serveFileWithPlaceholders serves out files that contain placeholders, ie. HTML and JS files
func serveFileWithPlaceholders(filePath string, w http.ResponseWriter) {
	logger.Debugf("Serving file %s replacing placeholders", filePath)

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

// staticHandler returns a middleware that serves the static content of the app, which includes the stuff listed in
// UIHTMLPaths[] and UIStaticPaths[] (favicon, scripts and such)
func staticHandler(next http.Handler) http.Handler {
	// Instantiate a file server for static content
	fileHandler := http.FileServer(http.Dir(config.CLIFlags.StaticPath))

	// Make a middleware handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok, p := config.PathOfBaseURL(r.URL.Path); ok && r.Method == "GET" {
			// Check if it's an HTML page
			if util.UIHTMLPaths[p] {
				serveFileWithPlaceholders(fmt.Sprintf("/html/%s.html", p), w)
				return
			}

			// Check if it's an other file with placeholders
			if util.UIPlaceholderPaths[p] {
				serveFileWithPlaceholders(p, w)
				return
			}

			// Check if it's a static file. Do not allow directory browsing (ie. paths ending with a '/')
			if util.UIStaticPaths[p] && !strings.HasSuffix(p, "/") {
				logger.Debugf("Serving static file /%s", p)

				// Make a "fake" (bypassing) response writer
				bypassWriter := &notFoundBypassWriter{ResponseWriter: w}

				// Try to serve the requested static content
				r.URL.Path = "/" + p
				fileHandler.ServeHTTP(bypassWriter, r)

				// If the content was found, we're done
				if bypassWriter.status != http.StatusNotFound {
					return
				}
			}

			// Remove any existing header to allow automatic MIME type detection
			delete(w.Header(), "Content-Type")
		}

		// Pass on to the next handler otherwise
		next.ServeHTTP(w, r)
	})
}
