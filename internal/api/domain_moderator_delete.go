package api

import (
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
)

func domainModeratorDelete(domain string, email string) error {
	if domain == "" || email == "" {
		return util.ErrorMissingConfig
	}

	statement := `delete from moderators where domain=$1 and email=$2;`
	_, err := svc.DB.Exec(statement, domain, email)
	if err != nil {
		logger.Errorf("cannot delete moderator: %v", err)
		return util.ErrorInternal
	}

	return nil
}

func domainModeratorDeleteHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
		Domain     *string `json:"domain"`
		Email      *string `json:"email"`
	}

	var x request
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	o, err := ownerGetByOwnerToken(*x.OwnerToken)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	domain := domainStrip(*x.Domain)
	authorised, err := domainOwnershipVerify(o.OwnerHex, domain)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if !authorised {
		BodyMarshalChecked(w, response{"success": false, "message": util.ErrorNotAuthorised.Error()})
		return
	}

	if err = domainModeratorDelete(domain, *x.Email); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true})
}
