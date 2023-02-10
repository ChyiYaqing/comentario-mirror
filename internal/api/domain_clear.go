package api

import (
	"gitlab.com/commento/commento/api/internal/util"
	"net/http"
)

func domainClear(domain string) error {
	if domain == "" {
		return util.ErrorMissingField
	}

	statement := `delete from votes using comments where comments.commentHex = votes.commentHex and comments.domain = $1;`
	_, err := DB.Exec(statement, domain)
	if err != nil {
		logger.Errorf("cannot delete votes: %v", err)
		return util.ErrorInternal
	}

	statement = `delete from comments where comments.domain = $1;`
	_, err = DB.Exec(statement, domain)
	if err != nil {
		logger.Errorf(statement, domain)
		return util.ErrorInternal
	}

	statement = `delete from pages where pages.domain = $1;`
	_, err = DB.Exec(statement, domain)
	if err != nil {
		logger.Errorf(statement, domain)
		return util.ErrorInternal
	}

	return nil
}

func domainClearHandler(w http.ResponseWriter, r *http.Request) {
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

	if err = domainClear(*x.Domain); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true})
}
