package config

import (
	"github.com/op/go-logging"
	"gitlab.com/comentario/comentario/internal/util"
	"testing"
)

func FailTestOnError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("failed test: %v", err)
	}
}

var setupComplete bool

func SetupTestEnv() error {
	if !setupComplete {
		setupComplete = true

		if err := util.LoggerCreate(); err != nil {
			return err
		}

		// Print messages to console only if verbose. Sounds like a good idea to
		// keep the console clean on `go test`.
		if !testing.Verbose() {
			logging.SetLevel(logging.CRITICAL, "")
		}
	}
	return nil
}
