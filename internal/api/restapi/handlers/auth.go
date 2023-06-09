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
		if owner, err := svc.TheUserService.FindOwnerByEmail(email, false); err == nil {
			user = &owner.User
		} else if err != svc.ErrNotFound {
			return respServiceError(err)
		}

	// Resetting commenter password: find the locally authenticated commenter
	case models.EntityCommenter:
		if commenter, err := svc.TheUserService.FindCommenterByIdPEmail("", email, false); err == nil {
			user = &commenter.User
		} else if err != svc.ErrNotFound {
			return respServiceError(err)
		}
	}

	// If no user found, apply a random delay to discourage email polling
	if user == nil {
		util.RandomSleep(100*time.Millisecond, 4000*time.Millisecond)

		// Generate a random reset token
	} else if token, err := svc.TheUserService.CreateResetToken(user.HexID, entity); err != nil {
		return respServiceError(err)

		// Send out an email
	} else if err := svc.TheMailService.SendFromTemplate(
		"",
		email,
		"Reset your password",
		"reset-hex.gohtml",
		map[string]any{"URL": config.URLFor("reset", map[string]string{"hex": string(token)})},
	); err != nil {
		return respServiceError(err)
	}

	// Succeeded (or no user found)
	return operations.NewForgotPasswordNoContent()
}

func ResetPassword(params operations.ResetPasswordParams) middleware.Responder {
	entity, err := svc.TheUserService.ResetUserPasswordByToken(*params.Body.ResetHex, *params.Body.Password)
	if err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewResetPasswordOK().WithPayload(&operations.ResetPasswordOKBody{Entity: entity})
}
