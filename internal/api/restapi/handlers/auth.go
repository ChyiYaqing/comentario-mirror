package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func ForgotPassword(params operations.ForgotPasswordParams) middleware.Responder {
	if err := forgotPassword(*params.Body.Email, *params.Body.Entity); err != nil {
		return operations.NewForgotPasswordOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	// Succeeded
	return operations.NewForgotPasswordOK().WithPayload(&models.APIResponseBase{Success: true})
}

func ResetPassword(params operations.ResetPasswordParams) middleware.Responder {
	entity, err := resetPassword(*params.Body.ResetHex, *params.Body.Password)
	if err != nil {
		return operations.NewResetPasswordOK().WithPayload(&operations.ResetPasswordOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewResetPasswordOK().WithPayload(&operations.ResetPasswordOKBody{
		Entity:  entity,
		Success: true,
	})
}

func forgotPassword(email strfmt.Email, entity models.Entity) error {
	if email == "" {
		return util.ErrorMissingField
	}

	var hex models.HexID
	if entity == models.EntityOwner {
		user, err := svc.TheUserService.FindOwnerByEmail(string(email))
		if err != nil {
			if err == util.ErrorNoSuchEmail {
				// TODO: use a more random time instead.
				time.Sleep(1 * time.Second)
				return nil
			} else {
				logger.Errorf("cannot get owner by email: %v", err)
				return util.ErrorInternal
			}
		}
		hex = user.HexID
	} else {
		commenter, err := commenterGetByEmail("commento", email)
		if err != nil {
			if err == util.ErrorNoSuchEmail {
				// TODO: use a more random time instead.
				time.Sleep(1 * time.Second)
				return nil
			} else {
				logger.Errorf("cannot get commenter by email: %v", err)
				return util.ErrorInternal
			}
		}
		hex = models.HexID(commenter.CommenterHex)
	}

	resetHex, err := util.RandomHex(32)
	if err != nil {
		return err
	}

	_, err = svc.DB.Exec(
		"insert into resetHexes(resetHex, hex, entity, sendDate) values($1, $2, $3, $4);",
		resetHex,
		hex,
		entity,
		time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert resetHex: %v", err)
		return util.ErrorInternal
	}

	return svc.TheEmailService.SendFromTemplate(
		"",
		string(email),
		"Reset your password",
		"reset-hex.gohtml",
		map[string]any{"URL": config.URLFor("reset", map[string]string{"hex": resetHex})})
}

func resetPassword(resetHex models.HexID, password string) (models.Entity, error) {
	if resetHex == "" || password == "" {
		return "", util.ErrorMissingField
	}

	row := svc.DB.QueryRow("select hex, entity from resetHexes where resetHex = $1;", resetHex)

	var hex string
	var entity models.Entity
	if err := row.Scan(&hex, &entity); err != nil {
		// TODO: is this the only error?
		return "", util.ErrorNoSuchResetToken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("cannot generate hash from password: %v\n", err)
		return "", util.ErrorInternal
	}

	var statement string
	if entity == models.EntityOwner {
		statement = "update owners set passwordHash = $1 where ownerHex = $2;"
	} else {
		statement = "update commenters set passwordHash = $1 where commenterHex = $2;"
	}

	_, err = svc.DB.Exec(statement, string(passwordHash), hex)
	if err != nil {
		logger.Errorf("cannot change %s's password: %v\n", entity, err)
		return "", util.ErrorInternal
	}

	if _, err = svc.DB.Exec("delete from resetHexes where resetHex = $1;", resetHex); err != nil {
		logger.Warningf("cannot remove resetHex: %v\n", err)
	}
	return entity, nil
}
