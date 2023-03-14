package restapi

import (
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
)

// FindOwnerByCookieHeader determines if the owner token contained in the cookie, extracted from the passed Cookie
// header, checks out
func FindOwnerByCookieHeader(headerValue string) (*data.User, error) {
	// Hack to parse the provided data (which is in fact the "Cookie" header, but Swagger 2.0 doesn't support
	// auth cookies, only headers)
	r := http.Request{Header: http.Header{"Cookie": []string{headerValue}}}

	// Check if there's a session cookie
	if cookie, err := r.Cookie(util.UserTokenName); err == nil {
		if user, err := svc.TheUserService.FindOwnerByToken(models.HexID(cookie.Value)); err == nil {
			return &user.User, nil
		}
	}

	// Authentication failed
	return nil, ErrUnauthorised
}
