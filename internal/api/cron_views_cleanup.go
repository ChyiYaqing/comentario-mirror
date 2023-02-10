package api

import (
	"time"
)

func ViewsCleanupBegin() error {
	go func() {
		for {
			statement := `
				delete from views
				where viewDate < $1;
			`
			_, err := DB.Exec(statement, time.Now().UTC().AddDate(0, 0, -45))
			if err != nil {
				logger.Errorf("error cleaning up views: %v", err)
				return
			}

			time.Sleep(24 * time.Hour)
		}
	}()

	return nil
}
