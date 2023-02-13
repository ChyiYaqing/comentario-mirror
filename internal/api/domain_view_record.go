package api

import (
	"gitlab.com/comentario/comentario/internal/svc"
	"time"
)

func domainViewRecord(domain string, commenterHex string) {
	statement := `
		insert into
		views  (domain, commenterHex, viewDate)
		values ($1,     $2,           $3      );
	`
	_, err := svc.DB.Exec(statement, domain, commenterHex, time.Now().UTC())
	if err != nil {
		logger.Warningf("cannot insert views: %v", err)
	}
}
