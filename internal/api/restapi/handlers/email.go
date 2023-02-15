package handlers

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"time"
)

const emailsRowColumns = `
	emails.email,
	emails.unsubscribeSecretHex,
	emails.lastEmailNotificationDate,
	emails.sendReplyNotifications,
	emails.sendModeratorNotifications
`

func EmailGet(params operations.EmailGetParams) middleware.Responder {
	email, err := emailGetByUnsubscribeSecretHex(*params.Body.UnsubscribeSecretHex)
	if err != nil {
		return operations.NewEmailGetOK().WithPayload(&operations.EmailGetOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewEmailGetOK().WithPayload(&operations.EmailGetOKBody{
		Email:   email,
		Success: true,
	})
}

func EmailNew(email strfmt.Email) error {
	unsubscribeSecretHex, err := util.RandomHex(32)
	if err != nil {
		return util.ErrorInternal
	}

	_, err = svc.DB.Exec(
		`insert into emails(email, unsubscribeSecretHex, lastEmailNotificationDate) values ($1, $2, $3) on conflict do nothing;`,
		email,
		unsubscribeSecretHex,
		time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert email into emails: %v", err)
		return util.ErrorInternal
	}

	return nil
}

func emailGet(em strfmt.Email) (*models.Email, error) {
	row := svc.DB.QueryRow(
		fmt.Sprintf("select %s from emails where email = $1;", emailsRowColumns),
		em)

	var e models.Email
	if err := emailsRowScan(row, &e); err != nil {
		// TODO: is this the only error?
		return nil, util.ErrorNoSuchEmail
	}

	return &e, nil
}

func emailGetByUnsubscribeSecretHex(unsubscribeSecretHex models.HexID) (*models.Email, error) {
	row := svc.DB.QueryRow(
		fmt.Sprintf("select %s from emails where unsubscribesecrethex = $1;", emailsRowColumns),
		unsubscribeSecretHex)

	var e models.Email
	if err := emailsRowScan(row, &e); err != nil {
		// TODO: is this the only error?
		return nil, util.ErrorNoSuchUnsubscribeSecretHex
	}

	return &e, nil
}

func emailsRowScan(s util.Scanner, e *models.Email) error {
	return s.Scan(
		&e.Email,
		&e.UnsubscribeSecretHex,
		&e.LastEmailNotificationDate,
		&e.SendReplyNotifications,
		&e.SendModeratorNotifications,
	)
}
