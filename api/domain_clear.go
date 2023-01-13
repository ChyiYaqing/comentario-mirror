package main

import (
	"net/http"
)

func domainClear(domain string) error {
	if domain == "" {
		return errorMissingField
	}

	statement := `delete from votes using comments where comments.commentHex = votes.commentHex and comments.domain = $1;`
	_, err := db.Exec(statement, domain)
	if err != nil {
		logger.Errorf("cannot delete votes: %v", err)
		return errorInternal
	}

	statement = `delete from comments where comments.domain = $1;`
	_, err = db.Exec(statement, domain)
	if err != nil {
		logger.Errorf(statement, domain)
		return errorInternal
	}

	statement = `delete from pages where pages.domain = $1;`
	_, err = db.Exec(statement, domain)
	if err != nil {
		logger.Errorf(statement, domain)
		return errorInternal
	}

	return nil
}

func domainClearHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
		Domain     *string `json:"domain"`
	}

	var x request
	if err := bodyUnmarshal(r, &x); err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	o, err := ownerGetByOwnerToken(*x.OwnerToken)
	if err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	domain := domainStrip(*x.Domain)
	isOwner, err := domainOwnershipVerify(o.OwnerHex, domain)
	if err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if !isOwner {
		bodyMarshalChecked(w, response{"success": false, "message": errorNotAuthorised.Error()})
		return
	}

	if err = domainClear(*x.Domain); err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	bodyMarshalChecked(w, response{"success": true})
}
