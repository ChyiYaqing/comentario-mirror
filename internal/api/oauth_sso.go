package api

import (
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"time"
)

type ssoPayload struct {
	Domain string `json:"domain"`
	Token  string `json:"token"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Link   string `json:"link"`
	Photo  string `json:"photo"`
}

func ssoTokenNew(domain string, commenterToken string) (string, error) {
	token, err := util.RandomHex(32)
	if err != nil {
		logger.Errorf("error generating SSO token hex: %v", err)
		return "", util.ErrorInternal
	}

	statement := `
		insert into
		ssoTokens (token, domain, commenterToken, creationDate)
		values    ($1,    $2,     $3,             $4          );
	`
	_, err = svc.DB.Exec(statement, token, domain, commenterToken, time.Now().UTC())
	if err != nil {
		logger.Errorf("error inserting SSO token: %v", err)
		return "", util.ErrorInternal
	}

	return token, nil
}

func ssoTokenExtract(token string) (string, string, error) {
	statement := "select domain, commenterToken from ssoTokens where token = $1;"
	row := svc.DB.QueryRow(statement, token)

	var domain string
	var commenterToken string
	if err := row.Scan(&domain, &commenterToken); err != nil {
		return "", "", util.ErrorNoSuchToken
	}

	statement = `
		delete from ssoTokens
		where token = $1;
	`
	if _, err := svc.DB.Exec(statement, token); err != nil {
		logger.Errorf("cannot delete SSO token after usage: %v", err)
		return "", "", util.ErrorInternal
	}

	return domain, commenterToken, nil
}
