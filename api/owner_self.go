package main

import (
	"net/http"
)

func ownerSelfHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
	}

	var x request
	if err := bodyUnmarshal(r, &x); err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	o, err := ownerGetByOwnerToken(*x.OwnerToken)
	if err == errorNoSuchToken {
		bodyMarshalChecked(w, response{"success": true, "loggedIn": false})
		return
	}

	if err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	bodyMarshalChecked(w, response{"success": true, "loggedIn": true, "owner": o})
}
