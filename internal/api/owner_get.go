package api

import (
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
)

var ownersRowColumns = `
	owners.ownerHex,
	owners.email,
	owners.name,
	owners.confirmedEmail,
	owners.joinDate
`

func ownersRowScan(s sqlScanner, o *models.Owner) error {
	return s.Scan(
		&o.OwnerHex,
		&o.Email,
		&o.Name,
		&o.ConfirmedEmail,
		&o.JoinDate,
	)
}

func ownerGetByEmail(email strfmt.Email) (*models.Owner, error) {
	if email == "" {
		return nil, util.ErrorMissingField
	}

	statement := `
		SELECT ` + ownersRowColumns + `
		FROM owners
		WHERE email=$1;
	`
	row := svc.DB.QueryRow(statement, email)

	var o models.Owner
	if err := ownersRowScan(row, &o); err != nil {
		// TODO: Make sure this is actually no such email.
		return nil, util.ErrorNoSuchEmail
	}

	return &o, nil
}

func OwnerGetByOwnerToken(ownerToken models.HexID) (*models.Owner, error) {
	if ownerToken == "" {
		return nil, util.ErrorMissingField
	}

	statement := `
		SELECT ` + ownersRowColumns + `
		FROM owners
		WHERE owners.ownerHex IN (
			SELECT ownerSessions.ownerHex FROM ownerSessions
			WHERE ownerSessions.ownerToken = $1
		);
	`
	row := svc.DB.QueryRow(statement, ownerToken)

	var o models.Owner
	if err := ownersRowScan(row, &o); err != nil {
		logger.Errorf("cannot scan owner: %v\n", err)
		return nil, util.ErrorInternal
	}

	return &o, nil
}
