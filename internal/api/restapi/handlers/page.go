package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/svc"
)

func PageUpdate(params operations.PageUpdateParams, principal data.Principal) middleware.Responder {
	// Verify the commenter is authenticated
	if r := Verifier.PrincipalIsAuthenticated(principal); r != nil {
		return r
	}

	// Verify the user is a domain moderator
	page := params.Body.Page
	if r := Verifier.UserIsDomainModerator(principal.GetUser().Email, swag.StringValue(page.Domain)); r != nil {
		return r
	}

	// Insert or update the page
	if err := svc.ThePageService.UpsertByDomainPath(page); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewPageUpdateNoContent()
}
