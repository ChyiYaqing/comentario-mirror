package api

import (
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func reset(resetHex string, password string) (string, error) {
	if resetHex == "" || password == "" {
		return "", util.ErrorMissingField
	}

	statement := `select hex, entity from resetHexes where resetHex = $1;`
	row := svc.DB.QueryRow(statement, resetHex)

	var hex string
	var entity string
	if err := row.Scan(&hex, &entity); err != nil {
		// TODO: is this the only error?
		return "", util.ErrorNoSuchResetToken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("cannot generate hash from password: %v\n", err)
		return "", util.ErrorInternal
	}

	if entity == "owner" {
		statement = `update owners set passwordHash = $1 where ownerHex = $2;`
	} else {
		statement = `update commenters set passwordHash = $1 where commenterHex = $2;`
	}

	_, err = svc.DB.Exec(statement, string(passwordHash), hex)
	if err != nil {
		logger.Errorf("cannot change %s's password: %v\n", entity, err)
		return "", util.ErrorInternal
	}

	statement = `delete from resetHexes where resetHex = $1;`
	_, err = svc.DB.Exec(statement, resetHex)
	if err != nil {
		logger.Warningf("cannot remove resetHex: %v\n", err)
	}

	return entity, nil
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		ResetHex *string `json:"resetHex"`
		Password *string `json:"password"`
	}

	var x request
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	entity, err := reset(*x.ResetHex, *x.Password)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true, "entity": entity})
}
