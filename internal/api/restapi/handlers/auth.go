package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"time"
)

func ForgotPassword(params operations.ForgotPasswordParams) middleware.Responder {
	email := data.EmailToString(params.Body.Email)
	entity := *params.Body.Entity

	var user *data.User

	switch entity {
	// Resetting owner password
	case models.EntityOwner:
		if owner, err := svc.TheUserService.FindOwnerByEmail(email); err == nil {
			user = &owner.User
		} else if err != svc.ErrNotFound {
			return serviceErrorResponder(err)
		}

	// Resetting commenter password: find the locally authenticated commenter
	case models.EntityCommenter:
		if commenter, err := svc.TheUserService.FindCommenterByIdPEmail("", email); err == nil {
			user = &commenter.User
		} else if err != svc.ErrNotFound {
			return serviceErrorResponder(err)
		}
	}

	// If no user found, apply a random delay to discourage email polling
	if user == nil {
		util.RandomSleep(100*time.Millisecond, 4000*time.Millisecond)

		// Generate a random reset token
	} else if token, err := svc.TheUserService.CreateResetToken(user.HexID, entity); err != nil {
		return serviceErrorResponder(err)

		// Send out an email
	} else if err := svc.TheMailService.SendFromTemplate(
		"",
		email,
		"Reset your password",
		"reset-hex.gohtml",
		map[string]any{"URL": config.URLFor("reset", map[string]string{"hex": token})},
	); err != nil {
		return serviceErrorResponder(err)
	}

	// Succeeded (or no user found)
	return operations.NewForgotPasswordOK().WithPayload(&models.APIResponseBase{Success: true})
}

func ResetPassword(params operations.ResetPasswordParams) middleware.Responder {
	entity, err := svc.TheUserService.ResetUserPasswordByToken(*params.Body.ResetHex, *params.Body.Password)
	if err != nil {
		return serviceErrorResponder(err)
	}

	// Succeeded
	return operations.NewResetPasswordOK().WithPayload(&operations.ResetPasswordOKBody{
		Entity:  entity,
		Success: true,
	})
}
