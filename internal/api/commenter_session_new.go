package api

import (
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
	"time"
)

func commenterTokenNew() (string, error) {
	commenterToken, err := util.RandomHex(32)
	if err != nil {
		logger.Errorf("cannot create commenterToken: %v", err)
		return "", util.ErrorInternal
	}

	statement := `insert into commenterSessions(commenterToken, creationDate) values($1, $2);`
	_, err = svc.DB.Exec(statement, commenterToken, time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert new commenterToken: %v", err)
		return "", util.ErrorInternal
	}

	return commenterToken, nil
}

func commenterTokenNewHandler(w http.ResponseWriter, _ *http.Request) {
	commenterToken, err := commenterTokenNew()
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true, "commenterToken": commenterToken})
}
