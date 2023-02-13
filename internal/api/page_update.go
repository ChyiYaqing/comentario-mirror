package api

import (
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
)

func pageUpdate(p page) error {
	if p.Domain == "" {
		return util.ErrorMissingField
	}

	// fields to not update:
	//   commentCount
	statement := `
		insert into pages(domain, path, isLocked, stickyCommentHex) values($1, $2, $3, $4)
		on conflict (domain, path) do
			update set isLocked = $3, stickyCommentHex = $4;
	`
	_, err := DB.Exec(statement, p.Domain, p.Path, p.IsLocked, p.StickyCommentHex)
	if err != nil {
		logger.Errorf("error setting page attributes: %v", err)
		return util.ErrorInternal
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
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	c, err := commenterGetByCommenterToken(*x.CommenterToken)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	domain := domainStrip(*x.Domain)

	isModerator, err := isDomainModerator(domain, c.Email)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if !isModerator {
		BodyMarshalChecked(w, response{"success": false, "message": util.ErrorNotModerator.Error()})
		return
	}

	(*x.Attributes).Domain = *x.Domain
	(*x.Attributes).Path = *x.Path

	if err = pageUpdate(*x.Attributes); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true})
}
