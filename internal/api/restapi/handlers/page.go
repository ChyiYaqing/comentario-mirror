package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/svc"
)

func PageUpdate(params operations.PageUpdateParams) middleware.Responder {
	// Find the commenter
	commenter, err := svc.TheUserService.FindCommenterByToken(*params.Body.CommenterToken)
	if err != nil {
		return respServiceError(err)
	}

	// Verify the user is a moderator
	if isModerator, err := svc.TheDomainService.IsDomainModerator(*params.Body.Domain, commenter.Email); err != nil {
		return respServiceError(err)
	} else if !isModerator {
		return operations.NewGenericForbidden()
	}

	// Insert or update the page
	_, err = svc.ThePageService.UpsertByDomainPath(
		*params.Body.Domain,
		params.Body.Path,
		params.Body.Attributes.IsLocked,
		params.Body.Attributes.StickyCommentHex)
	if err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewPageUpdateOK().WithPayload(&models.APIResponseBase{Success: true})
}
