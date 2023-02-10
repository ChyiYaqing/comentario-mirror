package api

import (
	"gitlab.com/commento/commento/api/internal/util"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

func ownerLogin(email string, password string) (string, error) {
	if email == "" || password == "" {
		return "", util.ErrorMissingField
	}

	statement := `select ownerHex, confirmedEmail, passwordHash from owners where email=$1;`
	row := DB.QueryRow(statement, email)

	var ownerHex string
	var confirmedEmail bool
	var passwordHash string
	if err := row.Scan(&ownerHex, &confirmedEmail, &passwordHash); err != nil {
		return "", util.ErrorInvalidEmailPassword
	}

	if !confirmedEmail {
		return "", util.ErrorUnconfirmedEmail
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		// TODO: is this the only possible error?
		return "", util.ErrorInvalidEmailPassword
	}

	ownerToken, err := util.RandomHex(32)
	if err != nil {
		logger.Errorf("cannot create ownerToken: %v", err)
		return "", util.ErrorInternal
	}

	statement = `insert into ownerSessions(ownerToken, ownerHex, loginDate) values($1, $2, $3);`
	_, err = DB.Exec(statement, ownerToken, ownerHex, time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert ownerSession: %v\n", err)
		return "", util.ErrorInternal
	}

	return ownerToken, nil
}

func ownerLoginHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}

	var x request
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	ownerToken, err := ownerLogin(*x.Email, *x.Password)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true, "ownerToken": ownerToken})
}
