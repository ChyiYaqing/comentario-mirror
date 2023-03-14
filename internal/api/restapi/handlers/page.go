package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
)

func PageUpdate(params operations.PageUpdateParams) middleware.Responder {
	// Find the commenter
	commenter, err := svc.TheUserService.FindCommenterByToken(*params.Body.CommenterToken)
	if err != nil {
		return serviceErrorResponder(err)
	}

	isModerator, err := isDomainModerator(*params.Body.Domain, strfmt.Email(commenter.Email))
	if err != nil {
		return operations.NewPageUpdateOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	if !isModerator {
		return operations.NewPageUpdateOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotModerator.Error()})
	}

	// Insert or update the page
	_, err = svc.ThePageService.UpsertByDomainPath(
		*params.Body.Domain,
		params.Body.Path,
		params.Body.Attributes.IsLocked,
		params.Body.Attributes.StickyCommentHex)
	if err != nil {
		return serviceErrorResponder(err)
	}

	// Succeeded
	return operations.NewPageUpdateOK().WithPayload(&models.APIResponseBase{Success: true})
}
