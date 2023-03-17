package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/svc"
)

func PageUpdate(params operations.PageUpdateParams) middleware.Responder {
	// Find the commenter
	commenter, err := svc.TheUserService.FindCommenterByToken(*params.Body.CommenterToken)
	if err != nil {
		return respServiceError(err)
	}

	// Verify the user is a domain moderator
	domain := data.TrimmedString(params.Body.Domain)
	if r := Verifier.UserIsDomainModerator(commenter.Email, domain); r != nil {
		return r
	}

	// Insert or update the page
	_, err = svc.ThePageService.UpsertByDomainPath(
		domain,
		params.Body.Path,
		params.Body.Attributes.IsLocked,
		params.Body.Attributes.StickyCommentHex)
	if err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewPageUpdateNoContent()
}
