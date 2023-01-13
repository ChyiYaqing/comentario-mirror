package main

import (
	"net/http"
	"time"
)

func commenterTokenNew() (string, error) {
	commenterToken, err := randomHex(32)
	if err != nil {
		logger.Errorf("cannot create commenterToken: %v", err)
		return "", errorInternal
	}

	statement := `insert into commenterSessions(commenterToken, creationDate) values($1, $2);`
	_, err = db.Exec(statement, commenterToken, time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert new commenterToken: %v", err)
		return "", errorInternal
	}

	return commenterToken, nil
}

func commenterTokenNewHandler(w http.ResponseWriter, _ *http.Request) {
	commenterToken, err := commenterTokenNew()
	if err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	bodyMarshalChecked(w, response{"success": true, "commenterToken": commenterToken})
}
