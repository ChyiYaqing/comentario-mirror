package api

import (
	"gitlab.com/comentario/comentario/internal/svc"
	"time"
)

func ViewsCleanupBegin() error {
	go func() {
		for {
			statement := `delete from views where viewDate < $1;`
			_, err := svc.DB.Exec(statement, time.Now().UTC().AddDate(0, 0, -45))
			if err != nil {
				logger.Errorf("error cleaning up views: %v", err)
				return
			}

			time.Sleep(24 * time.Hour)
		}
	}()

	return nil
}
