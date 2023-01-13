package main

import (
	"net/http"
)

func pageUpdate(p page) error {
	if p.Domain == "" {
		return errorMissingField
	}

	// fields to not update:
	//   commentCount
	statement := `
		insert into pages(domain, path, isLocked, stickyCommentHex) values($1, $2, $3, $4)
		on conflict (domain, path) do
			update set isLocked = $3, stickyCommentHex = $4;
	`
	_, err := db.Exec(statement, p.Domain, p.Path, p.IsLocked, p.StickyCommentHex)
	if err != nil {
		logger.Errorf("error setting page attributes: %v", err)
		return errorInternal
	}

	return nil
}

func pageUpdateHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		CommenterToken *string `json:"commenterToken"`
		Domain         *string `json:"domain"`
		Path           *string `json:"path"`
		Attributes     *page   `json:"attributes"`
	}

	var x request
	if err := bodyUnmarshal(r, &x); err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	c, err := commenterGetByCommenterToken(*x.CommenterToken)
	if err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	domain := domainStrip(*x.Domain)

	isModerator, err := isDomainModerator(domain, c.Email)
	if err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if !isModerator {
		bodyMarshalChecked(w, response{"success": false, "message": errorNotModerator.Error()})
		return
	}

	(*x.Attributes).Domain = *x.Domain
	(*x.Attributes).Path = *x.Path

	if err = pageUpdate(*x.Attributes); err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	bodyMarshalChecked(w, response{"success": true})
}
