package api

import (
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
)

func ownerSelfHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
	}

	var x request
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	o, err := ownerGetByOwnerToken(*x.OwnerToken)
	if err == util.ErrorNoSuchToken {
		BodyMarshalChecked(w, response{"success": true, "loggedIn": false})
		return
	}

	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true, "loggedIn": true, "owner": o})
}
