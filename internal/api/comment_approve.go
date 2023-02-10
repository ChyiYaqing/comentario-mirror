package api

import (
	"gitlab.com/commento/commento/api/internal/util"
	"net/http"
)

func commentApprove(commentHex string) error {
	if commentHex == "" {
		return util.ErrorMissingField
	}

	statement := `update comments set state = 'approved' where commentHex = $1;`

	_, err := DB.Exec(statement, commentHex)
	if err != nil {
		logger.Errorf("cannot approve comment: %v", err)
		return util.ErrorInternal
	}

	return nil
}

func commentApproveHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		CommenterToken *string `json:"commenterToken"`
		CommentHex     *string `json:"commentHex"`
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

	domain, _, err := commentDomainPathGet(*x.CommentHex)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	isModerator, err := isDomainModerator(domain, c.Email)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if !isModerator {
		BodyMarshalChecked(w, response{"success": false, "message": util.ErrorNotModerator.Error()})
		return
	}

	if err = commentApprove(*x.CommentHex); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true})
}
