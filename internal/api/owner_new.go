package api

import (
	"gitlab.com/comentario/comentario/internal/mail"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"time"
)

func ownerNew(email string, name string, password string) (string, error) {
	if email == "" || name == "" || password == "" {
		return "", util.ErrorMissingField
	}

	if os.Getenv("FORBID_NEW_OWNERS") == "true" {
		return "", util.ErrorNewOwnerForbidden
	}

	if _, err := ownerGetByEmail(email); err == nil {
		return "", util.ErrorEmailAlreadyExists
	}

	if err := EmailNew(email); err != nil {
		return "", util.ErrorInternal
	}

	ownerHex, err := util.RandomHex(32)
	if err != nil {
		logger.Errorf("cannot generate ownerHex: %v", err)
		return "", util.ErrorInternal
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("cannot generate hash from password: %v\n", err)
		return "", util.ErrorInternal
	}

	statement := `insert into owners(ownerHex, email, name, passwordHash, joinDate, confirmedEmail) values($1, $2, $3, $4, $5, $6);`
	_, err = svc.DB.Exec(statement, ownerHex, email, name, string(passwordHash), time.Now().UTC(), !mail.SMTPConfigured)
	if err != nil {
		// TODO: Make sure `err` is actually about conflicting UNIQUE, and not some
		// other error. If it is something else, we should probably return `errorInternal`.
		return "", util.ErrorEmailAlreadyExists
	}

	if mail.SMTPConfigured {
		confirmHex, err := util.RandomHex(32)
		if err != nil {
			logger.Errorf("cannot generate confirmHex: %v", err)
			return "", util.ErrorInternal
		}

		statement = `
			insert into
			ownerConfirmHexes (confirmHex, ownerHex, sendDate)
			values            ($1,         $2,       $3      );
		`
		_, err = svc.DB.Exec(statement, confirmHex, ownerHex, time.Now().UTC())
		if err != nil {
			logger.Errorf("cannot insert confirmHex: %v\n", err)
			return "", util.ErrorInternal
		}

		if err = mail.SMTPOwnerConfirmHex(email, name, confirmHex); err != nil {
			return "", err
		}
	}

	return ownerHex, nil
}

func ownerNewHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    *string `json:"email"`
		Name     *string `json:"name"`
		Password *string `json:"password"`
	}

	var x request
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if _, err := ownerNew(*x.Email, *x.Name, *x.Password); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	// Errors in creating a commenter account should not hold this up.
	_, _ = commenterNew(*x.Email, *x.Name, "undefined", "undefined", "commento", *x.Password)

	BodyMarshalChecked(w, response{"success": true, "confirmEmail": mail.SMTPConfigured})
}
