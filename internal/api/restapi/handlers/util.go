package handlers

import (
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/op/go-logging"
	"net/http"
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
