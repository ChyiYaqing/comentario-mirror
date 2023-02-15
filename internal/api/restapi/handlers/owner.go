package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"gitlab.com/comentario/comentario/internal/api"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/util"
)

func OwnerSelf(params operations.OwnerSelfParams) middleware.Responder {
	// Try to find the owner
	owner, err := api.OwnerGetByOwnerToken(*params.Body.OwnerToken)
	if err == util.ErrorNoSuchToken {
		return operations.NewOwnerSelfOK().WithPayload(&operations.OwnerSelfOKBody{Success: true})
	}

	if err != nil {
		return operations.NewOwnerSelfOK().WithPayload(&operations.OwnerSelfOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewOwnerSelfOK().WithPayload(&operations.OwnerSelfOKBody{
		LoggedIn: true,
		Owner:    owner,
		Success:  true,
	})
}
