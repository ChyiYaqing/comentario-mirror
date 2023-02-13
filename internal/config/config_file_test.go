package config

import (
	"gitlab.com/comentario/comentario/internal/util"
	"os"
	"testing"
)

func TestConfigFileLoadBasics(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	f, err := os.CreateTemp("", "comentario")
	if err != nil {
		t.Errorf("error creating temporary file: %v", err)
		return
	}

	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("error closing temporary file: %v", err)
			return
		}

		if err := os.Remove(f.Name()); err != nil {
			t.Errorf("error removing temporary file: %v", err)
			return
		}
	}()

	contents := `
		# Comentario port
		COMENTARIO_PORT=8000
		COMENTARIO_GZIP_STATIC=true
	`
	if _, err := f.Write([]byte(contents)); err != nil {
		t.Errorf("error writing to temporary file: %v", err)
		return
	}

	_ = os.Setenv("COMENTARIO_PORT", "9000")
	if err := configFileLoad(f.Name()); err != nil {
		t.Errorf("unexpected error loading config file: %v", err)
		return
	}

	if os.Getenv("COMENTARIO_PORT") != "9000" {
		t.Errorf("expected COMENTARIO_PORT=9000 got COMENTARIO_PORT=%s", os.Getenv("COMENTARIO_PORT"))
		return
	}

	if os.Getenv("COMENTARIO_GZIP_STATIC") != "true" {
		t.Errorf("expected COMENTARIO_GZIP_STATIC=true got COMENTARIO_GZIP_STATIC=%s", os.Getenv("COMENTARIO_GZIP_STATIC"))
		return
	}
}

func TestConfigFileLoadInvalid(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	f, err := os.CreateTemp("", "comentario")
	if err != nil {
		t.Errorf("error creating temporary file: %v", err)
		return
	}

	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("error closing temporary file: %v", err)
			return
		}

		if err := os.Remove(f.Name()); err != nil {
			t.Errorf("error removing temporary file: %v", err)
			return
		}
	}()

	contents := `
		COMENTARIO_PORT=8000
		INVALID_LINE
	`
	if _, err := f.Write([]byte(contents)); err != nil {
		t.Errorf("error writing to temporary file: %v", err)
		return
	}

	if err := configFileLoad(f.Name()); err != util.ErrorInvalidConfigFile {
		t.Errorf("expected err=%v got err=%v", util.ErrorInvalidConfigFile, err)
		return
	}
}
