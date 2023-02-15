package api

import (
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"time"
)

func EmailNew(email strfmt.Email) error {
	unsubscribeSecretHex, err := util.RandomHex(32)
	if err != nil {
		return util.ErrorInternal
	}

	statement := `insert into emails(email, unsubscribeSecretHex, lastEmailNotificationDate) values ($1,    $2,                   $3                       ) on conflict do nothing;`
	_, err = svc.DB.Exec(statement, email, unsubscribeSecretHex, time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert email into emails: %v", err)
		return util.ErrorInternal
	}

	return nil
}
