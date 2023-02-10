package main

import (
	"fmt"
	"github.com/op/go-logging"
	"gitlab.com/commento/commento/api/internal/api"
	"gitlab.com/commento/commento/api/internal/config"
	"gitlab.com/commento/commento/api/internal/mail"
	"gitlab.com/commento/commento/api/internal/util"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var logger = logging.MustGetLogger("main")

func main() {
	exitIfError(util.LoggerCreate())
	exitIfError(versionPrint())
	exitIfError(config.ConfigParse())
	exitIfError(api.DBConnect(10, time.Second))
	exitIfError(api.Migrate())
	exitIfError(mail.SMTPConfigure())
	exitIfError(mail.SMTPTemplatesLoad())
	exitIfError(api.OAuthConfigure())
	exitIfError(api.MarkdownRendererCreate())
	exitIfError(sigintCleanupSetup())
	exitIfError(config.VersionCheckStart())
	exitIfError(api.DomainExportCleanupBegin())
	exitIfError(api.ViewsCleanupBegin())
	exitIfError(api.SSOTokenCleanupBegin())
	exitIfError(api.RoutesServe())
}

func exitIfError(err error) {
	if err != nil {
		fmt.Printf("fatal error: %v\n", err)
		os.Exit(1)
	}
}

func versionPrint() error {
	logger.Infof("starting Comentario %s", config.Version)
	return nil
}

func sigintCleanup() int {
	if api.DB != nil {
		err := api.DB.Close()
		if err == nil {
			logger.Errorf("cannot close database connection: %v", err)
			return 1
		}
	}
	return 0
}

func sigintCleanupSetup() error {
	logger.Infof("setting up SIGINT cleanup")

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	go func() {
		<-c
		os.Exit(sigintCleanup())
	}()

	return nil
}
