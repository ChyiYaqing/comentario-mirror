package api

import (
	"gitlab.com/comentario/comentario/internal/svc"
	"time"
)

func SSOTokenCleanupBegin() error {
	go func() {
		for {
			statement := `
				delete from ssoTokens
				where creationDate < $1;
			`
			_, err := svc.DB.Exec(statement, time.Now().UTC().Add(time.Duration(-10)*time.Minute))
			if err != nil {
				logger.Errorf("error cleaning up export rows: %v", err)
				return
			}

			time.Sleep(10 * time.Minute)
		}
	}()

	return nil
}
