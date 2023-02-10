package api

import (
	"gitlab.com/commento/commento/api/internal/util"
	"net/http"
)

func commentEdit(commentHex string, markdown string) (string, error) {
	if commentHex == "" {
		return "", util.ErrorMissingField
	}

	html := markdownToHtml(markdown)

	statement := `update comments set markdown = $2, html = $3 where commentHex=$1;`
	_, err := DB.Exec(statement, commentHex, markdown, html)

	if err != nil {
		// TODO: make sure this is the error is actually nonexistent commentHex
		return "", util.ErrorNoSuchComment
	}

	return html, nil
}

func commentEditHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		CommenterToken *string `json:"commenterToken"`
		CommentHex     *string `json:"commentHex"`
		Markdown       *string `json:"markdown"`
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

	cm, err := commentGetByCommentHex(*x.CommentHex)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if cm.CommenterHex != c.CommenterHex {
		BodyMarshalChecked(w, response{"success": false, "message": util.ErrorNotAuthorised.Error()})
		return
	}

	html, err := commentEdit(*x.CommentHex, *x.Markdown)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true, "html": html})
}
