package restapi

import (
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/handlers"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
)

// AuthCommenterByTokenHeader determines if the commenter token, contained in the X-Commenter-Token header, checks out
func AuthCommenterByTokenHeader(headerValue string) (data.Principal, error) {
	// Validate the token format
	if token := models.HexID(headerValue); token.Validate(nil) == nil {
		// If it's an anonymous commenter
		if token == data.AnonymousCommenter.HexID {
			return &data.AnonymousCommenter, nil
		}

		// Try to find the commenter by that token
		if commenter, err := svc.TheUserService.FindCommenterByToken(token); err == nil {
			return commenter, nil
		}
	}

	// Authentication failed
	return nil, ErrUnauthorised
}

// AuthOwnerByCookieHeader determines if the owner token contained in the cookie, extracted from the passed Cookie
// header, checks out
func AuthOwnerByCookieHeader(headerValue string) (data.Principal, error) {
	// Hack to parse the provided data (which is in fact the "Cookie" header, but Swagger 2.0 doesn't support
	// auth cookies, only headers)
	r := &http.Request{Header: http.Header{"Cookie": []string{headerValue}}}

	// Check if there's a token cookie
	if token := handlers.ExtractOwnerTokenFromCookie(r); token != "" {
		if owner, err := svc.TheUserService.FindOwnerByToken(token); err == nil {
			return owner, nil
		}
	}

	// Check if there's a token cookie
	if cookie, err := r.Cookie(util.CookieNameUserToken); err == nil {
		if owner, err := svc.TheUserService.FindOwnerByToken(models.HexID(cookie.Value)); err == nil {
			return owner, nil
		}
	}

	// Authentication failed
	return nil, ErrUnauthorised
}
