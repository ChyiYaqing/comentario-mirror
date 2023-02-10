package api

import (
	"gitlab.com/commento/commento/api/internal/util"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

func commenterLogin(email string, password string) (string, error) {
	if email == "" || password == "" {
		return "", util.ErrorMissingField
	}

	statement := `select commenterHex, passwordHash from commenters where email = $1 and provider = 'commento';`
	row := DB.QueryRow(statement, email)

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

	statement = `insert into commenterSessions(commenterToken, commenterHex, creationDate) values($1, $2, $3);`
	_, err = DB.Exec(statement, commenterToken, commenterHex, time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert commenterToken token: %v\n", err)
		return "", util.ErrorInternal
	}

	return commenterToken, nil
}

func commenterLoginHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}

	var x request
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	commenterToken, err := commenterLogin(*x.Email, *x.Password)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	// TODO: modify commenterLogin to directly return c?
	c, err := commenterGetByCommenterToken(commenterToken)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	e, err := emailGet(c.Email)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true, "commenterToken": commenterToken, "commenter": c, "email": e})
}
