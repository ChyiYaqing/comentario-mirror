package util

import (
	"testing"
)

func TestLoggerCreateBasics(t *testing.T) {
	logger = nil

	if err := LoggerCreate(); err != nil {
		t.Errorf("unexpected error creating logger: %v", err)
		return
	}

	if logger == nil {
		t.Errorf("logger null after LoggerCreate()")
		return
	}

	logger.Debugf("test message please ignore")
}
