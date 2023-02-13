package api

import (
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
	"time"
)

func commentVote(commenterHex string, commentHex string, direction int) error {
	if commentHex == "" || commenterHex == "" {
		return util.ErrorMissingField
	}

	statement := `select commenterHex from comments where commentHex = $1;`
	row := DB.QueryRow(statement, commentHex)

	var authorHex string
	if err := row.Scan(&authorHex); err != nil {
		logger.Errorf("error selecting authorHex for vote")
		return util.ErrorInternal
	}

	if authorHex == commenterHex {
		return util.ErrorSelfVote
	}

	statement = `
		insert into votes(commentHex, commenterHex, direction, voteDate) values($1, $2, $3, $4)
		on conflict (commentHex, commenterHex) do
			update set direction = $3;
	`
	_, err := DB.Exec(statement, commentHex, commenterHex, direction, time.Now().UTC())
	if err != nil {
		logger.Errorf("error inserting/updating votes: %v", err)
		return util.ErrorInternal
	}

	return nil
}

func commentVoteHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		CommenterToken *string `json:"commenterToken"`
		CommentHex     *string `json:"commentHex"`
		Direction      *int    `json:"direction"`
	}

	var x request
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if *x.CommenterToken == "anonymous" {
		BodyMarshalChecked(w, response{"success": false, "message": util.ErrorUnauthorisedVote.Error()})
		return
	}

	c, err := commenterGetByCommenterToken(*x.CommenterToken)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	direction := 0
	if *x.Direction > 0 {
		direction = 1
	} else if *x.Direction < 0 {
		direction = -1
	}

	if err := commentVote(c.CommenterHex, *x.CommentHex, direction); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true})
}
