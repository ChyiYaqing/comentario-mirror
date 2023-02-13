package api

import (
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
)

func domainSsoSecretNew(domain string) (string, error) {
	if domain == "" {
		return "", util.ErrorMissingField
	}

	ssoSecret, err := util.RandomHex(32)
	if err != nil {
		logger.Errorf("error generating SSO secret hex: %v", err)
		return "", util.ErrorInternal
	}

	statement := `update domains set ssoSecret = $2 where domain = $1;`
	_, err = svc.DB.Exec(statement, domain, ssoSecret)
	if err != nil {
		logger.Errorf("cannot update ssoSecret: %v", err)
		return "", util.ErrorInternal
	}

	return ssoSecret, nil
}

func domainSsoSecretNewHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
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
	isOwner, err := domainOwnershipVerify(o.OwnerHex, domain)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if !isOwner {
		BodyMarshalChecked(w, response{"success": false, "message": util.ErrorNotAuthorised.Error()})
		return
	}

	ssoSecret, err := domainSsoSecretNew(domain)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true, "ssoSecret": ssoSecret})
}
