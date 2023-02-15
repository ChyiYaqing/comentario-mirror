package handlers

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/mail"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

const commentersRowColumns = `
	commenters.commenterHex,
	commenters.email,
	commenters.name,
	commenters.link,
	commenters.photo,
	commenters.provider,
	commenters.joinDate
`

func CommenterLogin(params operations.CommenterLoginParams) middleware.Responder {
	commenterToken, err := commenterLogin(*params.Body.Email, *params.Body.Password)
	if err != nil {
		return operations.NewCommenterLoginOK().WithPayload(&operations.CommenterLoginOKBody{Message: err.Error()})
	}

	// TODO: modify commenterLogin to directly return c?
	commenter, err := commenterGetByCommenterToken(commenterToken)
	if err != nil {
		return operations.NewCommenterLoginOK().WithPayload(&operations.CommenterLoginOKBody{Message: err.Error()})
	}

	email, err := emailGet(commenter.Email)
	if err != nil {
		return operations.NewCommenterLoginOK().WithPayload(&operations.CommenterLoginOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewCommenterLoginOK().WithPayload(&operations.CommenterLoginOKBody{
		Commenter:      commenter,
		CommenterToken: commenterToken,
		Email:          email,
		Success:        true,
	})
}

func CommenterNew(params operations.CommenterNewParams) middleware.Responder {
	website := strings.TrimSpace(params.Body.Website)

	// TODO this is awful
	if website == "" {
		website = "undefined"
	}

	if _, err := commenterNew(*params.Body.Email, *params.Body.Name, website, "undefined", "commento", *params.Body.Password); err != nil {
		return operations.NewCommenterNewOK().WithPayload(&operations.CommenterNewOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewCommenterNewOK().WithPayload(&operations.CommenterNewOKBody{
		ConfirmEmail: mail.SMTPConfigured,
		Success:      true,
	})
}

func commenterGetByCommenterToken(commenterToken models.CommenterToken) (*models.Commenter, error) {
	if commenterToken == "" {
		return nil, util.ErrorMissingField
	}

	row := svc.DB.QueryRow(
		fmt.Sprintf(
			"select %s from commentersessions "+
				"join commenters on commentersessions.commenterhex = commenters.commenterhex "+
				"where commentertoken = $1;",
			commentersRowColumns),
		commenterToken)

	var c models.Commenter
	if err := commentersRowScan(row, &c); err != nil {
		// TODO: is this the only error?
		return nil, util.ErrorNoSuchToken
	}

	if c.CommenterHex == "none" {
		return nil, util.ErrorNoSuchToken
	}

	return &c, nil
}

func commenterGetByEmail(provider string, email strfmt.Email) (*models.Commenter, error) {
	if provider == "" || email == "" {
		return nil, util.ErrorMissingField
	}
	row := svc.DB.QueryRow(
		fmt.Sprintf("select %s from commenters where email=$1 and provider=$2;", commentersRowColumns),
		email,
		provider,
	)

	var c models.Commenter
	if err := commentersRowScan(row, &c); err != nil {
		// TODO: is this the only error?
		return nil, util.ErrorNoSuchCommenter
	}

	return &c, nil
}

func commenterGetByHex(commenterHex string) (*models.Commenter, error) {
	if commenterHex == "" {
		return nil, util.ErrorMissingField
	}

	row := svc.DB.QueryRow(
		fmt.Sprintf("select %s from commenters where commenterHex = $1;", commentersRowColumns),
		commenterHex)
	var c models.Commenter
	if err := commentersRowScan(row, &c); err != nil {
		// TODO: is this the only error?
		return nil, util.ErrorNoSuchCommenter
	}

	return &c, nil
}

func commenterLogin(email strfmt.Email, password string) (models.CommenterToken, error) {
	if email == "" || password == "" {
		return "", util.ErrorMissingField
	}

	row := svc.DB.QueryRow(
		"select commenterHex, passwordHash from commenters where email=$1 and provider='commento';",
		email)

	var commenterHex string
	var passwordHash string
	if err := row.Scan(&commenterHex, &passwordHash); err != nil {
		return "", util.ErrorInvalidEmailPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		// TODO: is this the only possible error?
		return "", util.ErrorInvalidEmailPassword
	}

	commenterToken, err := util.RandomHex(32)
	if err != nil {
		logger.Errorf("cannot create commenterToken: %v", err)
		return "", util.ErrorInternal
	}

	_, err = svc.DB.Exec(
		"insert into commenterSessions(commenterToken, commenterHex, creationDate) values($1, $2, $3);",
		commenterToken,
		commenterHex,
		time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert commenterToken token: %v\n", err)
		return "", util.ErrorInternal
	}

	return models.CommenterToken(commenterToken), nil
}

func commenterNew(email strfmt.Email, name string, link string, photo string, provider string, password string) (models.HexID, error) {
	if email == "" || name == "" || link == "" || photo == "" || provider == "" {
		return "", util.ErrorMissingField
	}

	if provider == "commento" && password == "" {
		return "", util.ErrorMissingField
	}

	if link != "undefined" {
		if _, err := util.ParseAbsoluteURL(link); err != nil {
			return "", err
		}
	}

	if _, err := commenterGetByEmail(provider, email); err == nil {
		return "", util.ErrorEmailAlreadyExists
	}

	if err := EmailNew(email); err != nil {
		return "", util.ErrorInternal
	}

	commenterHex, err := util.RandomHex(32)
	if err != nil {
		return "", util.ErrorInternal
	}

	var passwordHash []byte
	if password != "" {
		passwordHash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			logger.Errorf("cannot generate hash from password: %v\n", err)
			return "", util.ErrorInternal
		}
	}

	statement := `insert into commenters(commenterHex, email, name, link, photo, provider, passwordHash, joinDate) values($1, $2, $3, $4, $5, $6, $7, $8);`
	_, err = svc.DB.Exec(statement, commenterHex, email, name, link, photo, provider, string(passwordHash), time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert commenter: %v", err)
		return "", util.ErrorInternal
	}

	return models.HexID(commenterHex), nil
}

func commentersRowScan(s util.Scanner, c *models.Commenter) error {
	return s.Scan(
		&c.CommenterHex,
		&c.Email,
		&c.Name,
		&c.Link,
		&c.Photo,
		&c.Provider,
		&c.JoinDate,
	)
}
