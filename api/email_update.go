package main

import (
	"net/http"
)

func emailUpdate(e email) error {
	statement := `
		update emails
		set sendReplyNotifications = $3, sendModeratorNotifications = $4
		where email = $1 and unsubscribeSecretHex = $2;
	`
	_, err := db.Exec(statement, e.Email, e.UnsubscribeSecretHex, e.SendReplyNotifications, e.SendModeratorNotifications)
	if err != nil {
		logger.Errorf("error updating email: %v", err)
		return errorInternal
	}

	return nil
}

func emailUpdateHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email *email `json:"email"`
	}

	var x request
	if err := bodyUnmarshal(r, &x); err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if err := emailUpdate(*x.Email); err != nil {
		bodyMarshalChecked(w, response{"success": true, "message": err.Error()})
		return
	}

	bodyMarshalChecked(w, response{"success": true})
}
