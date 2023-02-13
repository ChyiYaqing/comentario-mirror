package api

import (
	"gitlab.com/comentario/comentario/internal/mail"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
	"time"
)

func forgot(email string, entity string) error {
	if email == "" {
		return util.ErrorMissingField
	}

	if entity != "owner" && entity != "commenter" {
		return util.ErrorInvalidEntity
	}

	if !mail.SMTPConfigured {
		return util.ErrorSmtpNotConfigured
	}

	var hex string
	var name string
	if entity == "owner" {
		o, err := ownerGetByEmail(email)
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
		hex = o.OwnerHex
		name = o.Name
	} else {
		c, err := commenterGetByEmail("commento", email)
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
		hex = c.CommenterHex
		name = c.Name
	}

	resetHex, err := util.RandomHex(32)
	if err != nil {
		return err
	}

	var statement string

	statement = `insert into resetHexes(resetHex, hex, entity, sendDate) values($1, $2, $3, $4);`
	_, err = svc.DB.Exec(statement, resetHex, hex, entity, time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert resetHex: %v", err)
		return util.ErrorInternal
	}

	err = mail.SMTPResetHex(email, name, resetHex)
	if err != nil {
		return err
	}

	return nil
}

func forgotHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email  *string `json:"email"`
		Entity *string `json:"entity"`
	}

	var x request
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if err := forgot(*x.Email, *x.Entity); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true})
}
