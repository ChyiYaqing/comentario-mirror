package api

import (
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
	"strings"
	"time"
)

func domainNew(ownerHex string, name string, domain string) error {
	if ownerHex == "" || name == "" || domain == "" {
		return util.ErrorMissingField
	}

	if strings.Contains(domain, "/") {
		return util.ErrorInvalidDomain
	}

	statement := `insert into domains(ownerHex, name, domain, creationDate) values($1, $2, $3, $4);`
	_, err := svc.DB.Exec(statement, ownerHex, name, domain, time.Now().UTC())
	if err != nil {
		// TODO: Make sure this is really the error.
		return util.ErrorDomainAlreadyExists
	}

	return nil
}

func domainNewHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
		Name       *string `json:"name"`
		Domain     *string `json:"domain"`
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

	if err = domainNew(o.OwnerHex, *x.Name, domain); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if err = domainModeratorNew(domain, o.Email); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true, "domain": domain})
}
