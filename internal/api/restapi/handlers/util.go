package handlers

import (
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/op/go-logging"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/svc"
	"net/http"
	"time"
)

// logger represents a package-wide logger instance
var logger = logging.MustGetLogger("handlers")

// closeParentWindowResponse returns a responder that renders an HTML script closing the parent window
func closeParentWindowResponse() middleware.Responder {
	return NewHTMLResponder(http.StatusOK, "<html><script>window.parent.close()</script></html>")
}

//----------------------------------------------------------------------------------------------------------------------

// HTMLResponder is an implementation of middleware.Responder that serves out a static piece of HTML
type HTMLResponder struct {
	code int
	html string
}

// NewHTMLResponder creates HTMLResponder with default headers values
func NewHTMLResponder(code int, html string) *HTMLResponder {
	return &HTMLResponder{
		code: code,
		html: html,
	}
}

// WriteResponse to the client
func (r *HTMLResponder) WriteResponse(w http.ResponseWriter, _ runtime.Producer) {
	w.WriteHeader(r.code)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(r.html))
}

// ----------------------------------------------------------------------------------------------------------------------

// CookieResponder is an implementation of middleware.Responder that wraps another responder and sets the provided
// cookies before handing over to it
type CookieResponder struct {
	responder middleware.Responder
	cookies   map[string]*http.Cookie
}

// NewCookieResponder instantiates a new CookieResponder
func NewCookieResponder(responder middleware.Responder) *CookieResponder {
	return &CookieResponder{
		responder: responder,
		cookies:   make(map[string]*http.Cookie),
	}
}

func (r *CookieResponder) WriteResponse(rw http.ResponseWriter, p runtime.Producer) {
	// Add cookies to the response
	for _, c := range r.cookies {
		http.SetCookie(rw, c)
	}
	// Hand over to the original responder
	r.responder.WriteResponse(rw, p)
}

// WithCookie adds a new cookie to the response
func (r *CookieResponder) WithCookie(name, value, path string, maxAge time.Duration, httpOnly bool, sameSite http.SameSite) *CookieResponder {
	r.cookies[name] = &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   int(maxAge.Seconds()),
		Secure:   config.UseHTTPS,
		HttpOnly: httpOnly,
		SameSite: sameSite,
	}
	return r
}

// WithoutCookie removes a cookie in the response by submitting a "pre-expired" cookie
func (r *CookieResponder) WithoutCookie(name, path string) *CookieResponder {
	r.cookies[name] = &http.Cookie{Name: name, Path: path, MaxAge: -1}
	return r
}

// respBadRequest returns a responder that responds with HTTP Bad Request error
func respBadRequest(err error) middleware.Responder {
	return operations.NewGenericBadRequest().WithPayload(&operations.GenericBadRequestBody{Details: err.Error()})
}

// respInternalError returns a responder that responds with HTTP Internal Server Error
func respInternalError() middleware.Responder {
	return operations.NewGenericInternalServerError()
}

// respServiceError translates the provided error, returned by a service, into an appropriate error responder
func respServiceError(err error) middleware.Responder {
	switch err {
	case svc.ErrDuplicateEmail:
		return operations.NewGenericBadRequest().
			WithPayload(&operations.GenericBadRequestBody{Details: "Duplicate email"})
	case svc.ErrNotFound:
		return operations.NewGenericNotFound()
	case svc.ErrPageLocked:
		return operations.NewGenericBadRequest().
			WithPayload(&operations.GenericBadRequestBody{Details: "Unable to add comment: the page is locked"})
	}

	// Not recognised: return an internal error response
	return respInternalError()
}

// respUnauthorized returns a responder that responds with HTTP Unauthorized error
func respUnauthorized(err error) middleware.Responder {
	return operations.NewGenericUnauthorized().WithPayload(&operations.GenericUnauthorizedBody{Details: err.Error()})
}
