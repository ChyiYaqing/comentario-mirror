package api

import (
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
	"time"
)

func domainModeratorNew(domain string, email strfmt.Email) error {
	if domain == "" || email == "" {
		return util.ErrorMissingField
	}

	if err := EmailNew(email); err != nil {
		logger.Errorf("cannot create email when creating moderator: %v", err)
		return util.ErrorInternal
	}

	statement := `insert into moderators(domain, email, addDate) values($1, $2, $3);`
	_, err := svc.DB.Exec(statement, domain, email, time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert new moderator: %v", err)
		return util.ErrorInternal
	}

	return nil
}

func domainModeratorNewHandler(w http.ResponseWriter, r *http.Request) {
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

	o, err := OwnerGetByOwnerToken(models.HexID(*x.OwnerToken))
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	domain := domainStrip(*x.Domain)
	isOwner, err := domainOwnershipVerify(o.OwnerHex, domain)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if !isOwner {
		BodyMarshalChecked(w, response{"success": false, "message": util.ErrorNotAuthorised.Error()})
		return
	}

	if err = domainModeratorNew(domain, strfmt.Email(*x.Email)); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true})
}
