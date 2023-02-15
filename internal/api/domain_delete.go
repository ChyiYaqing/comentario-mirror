package api

import (
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
)

func domainDelete(domain string) error {
	if domain == "" {
		return util.ErrorMissingField
	}

	statement := `delete from domains where domain = $1;`
	_, err := svc.DB.Exec(statement, domain)
	if err != nil {
		return util.ErrorNoSuchDomain
	}

	statement = `delete from views where views.domain = $1;`
	_, err = svc.DB.Exec(statement, domain)
	if err != nil {
		logger.Errorf("cannot delete domain from views: %v", err)
		return util.ErrorInternal
	}

	statement = `delete from moderators where moderators.domain = $1;`
	_, err = svc.DB.Exec(statement, domain)
	if err != nil {
		logger.Errorf("cannot delete domain from moderators: %v", err)
		return util.ErrorInternal
	}

	statement = `delete from ssotokens where ssotokens.domain = $1;`
	_, err = svc.DB.Exec(statement, domain)
	if err != nil {
		logger.Errorf("cannot delete domain from ssotokens: %v", err)
		return util.ErrorInternal
	}

	// comments, votes, and pages are handled by domainClear
	if err = domainClear(domain); err != nil {
		logger.Errorf("cannot clear domain: %v", err)
		return util.ErrorInternal
	}

	return nil
}

func domainDeleteHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
		Domain     *string `json:"domain"`
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

	if err = domainDelete(*x.Domain); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true})
}
