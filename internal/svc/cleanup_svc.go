package svc

import "time"

var TheCleanupService CleanupService = &cleanupService{}

type CleanupService interface {
	Init() error
}

type cleanupService struct{}

func (s *cleanupService) Init() error {
	logger.Debugf("cleanupService: initialising")
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
	logger.Debugf("cleanupService: initialising domain export cleanup")
	go func() {
		for {
			if err := db.Exec("delete from exports where creationDate < $1;", time.Now().UTC().AddDate(0, 0, -7)); err != nil {
				logger.Errorf("cleanupService: error cleaning up domain export rows: %v", err)
				return
			}
			time.Sleep(2 * time.Hour)
		}
	}()

	return nil
}

func (s *cleanupService) ssoTokenCleanupBegin() error {
	logger.Debugf("cleanupService: initialising SSO token cleanup")
	go func() {
		for {
			if err := db.Exec("delete from ssoTokens where creationDate < $1;", time.Now().UTC().Add(time.Duration(-10)*time.Minute)); err != nil {
				logger.Errorf("cleanupService: error cleaning up SSO tokens: %v", err)
				return
			}
			time.Sleep(10 * time.Minute)
		}
	}()

	return nil
}

func (s *cleanupService) viewsCleanupBegin() error {
	logger.Debugf("cleanupService: initialising view stats cleanup")
	go func() {
		for {
			if err := db.Exec("delete from views where viewDate < $1;", time.Now().UTC().AddDate(0, 0, -45)); err != nil {
				logger.Errorf("cleanupService: error cleaning up view stats: %v", err)
				return
			}
			time.Sleep(24 * time.Hour)
		}
	}()

	return nil
}
