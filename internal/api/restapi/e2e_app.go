package restapi

import (
	"github.com/op/go-logging"
	"gitlab.com/comentario/comentario/internal/svc"
)

// e2eApp is an End2EndApp implementation, which links this app to the e2e plugin
type e2eApp struct {
	logger *logging.Logger
}

func (a *e2eApp) DBExec(query string, args ...any) error {
	_, e := svc.DB.Exec(query, args...)
	return e
}

func (a *e2eApp) DBInit() error {
	// Install migrations
	return svc.DB.Migrate()
}

func (a *e2eApp) LogError(fmt string, args ...any) {
	a.logger.Errorf(fmt, args...)
}

func (a *e2eApp) LogInfo(fmt string, args ...any) {
	a.logger.Infof(fmt, args...)
}

func (a *e2eApp) LogWarning(fmt string, args ...any) {
	a.logger.Warningf(fmt, args...)
}
