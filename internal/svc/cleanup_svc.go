package svc

import "time"

var TheCleanupService CleanupService = &cleanupService{}

type CleanupService interface {
	Init() error
}

type cleanupService struct{}

func (s *cleanupService) Init() error {
	if err := s.domainExportCleanupBegin(); err != nil {
		return err
	}
	if err := s.ssoTokenCleanupBegin(); err != nil {
		return err
	}
	if err := s.viewsCleanupBegin(); err != nil {
		return err
	}
	return nil
}

func (s *cleanupService) domainExportCleanupBegin() error {
	go func() {
		for {
			statement := `
				delete from exports
				where creationDate < $1;
			`
			_, err := DB.Exec(statement, time.Now().UTC().AddDate(0, 0, -7))
			if err != nil {
				logger.Errorf("error cleaning up export rows: %v", err)
				return
			}

			time.Sleep(2 * time.Hour)
		}
	}()

	return nil
}

func (s *cleanupService) ssoTokenCleanupBegin() error {
	go func() {
		for {
			statement := `
				delete from ssoTokens
				where creationDate < $1;
			`
			_, err := DB.Exec(statement, time.Now().UTC().Add(time.Duration(-10)*time.Minute))
			if err != nil {
				logger.Errorf("error cleaning up export rows: %v", err)
				return
			}

			time.Sleep(10 * time.Minute)
		}
	}()

	return nil
}

func (s *cleanupService) viewsCleanupBegin() error {
	go func() {
		for {
			statement := `delete from views where viewDate < $1;`
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
