package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigParseBasics(t *testing.T) {
	_ = os.Setenv("COMENTARIO_ORIGIN", "https://comentario.app")

	if err := configParse(); err != nil {
		t.Errorf("unexpected error when parsing config: %v", err)
		return
	}

	if os.Getenv("BIND_ADDRESS") != "127.0.0.1" {
		t.Errorf("expected COMENTARIO_BIND_ADDRESS=127.0.0.1, but COMENTARIO_BIND_ADDRESS=%s instead", os.Getenv("BIND_ADDRESS"))
		return
	}

	_ = os.Setenv("COMENTARIO_BIND_ADDRESS", "192.168.1.100")

	_ = os.Setenv("COMENTARIO_PORT", "")
	if err := configParse(); err != nil {
		t.Errorf("unexpected error when parsing config: %v", err)
		return
	}

	if os.Getenv("BIND_ADDRESS") != "192.168.1.100" {
		t.Errorf("expected COMENTARIO_BIND_ADDRESS=192.168.1.100, but COMENTARIO_BIND_ADDRESS=%s instead", os.Getenv("BIND_ADDRESS"))
		return
	}

	// This test feels kinda stupid, but whatever.
	if os.Getenv("PORT") != "8080" {
		t.Errorf("expected PORT=8080, but PORT=%s instead", os.Getenv("PORT"))
		return
	}

	_ = os.Setenv("COMENTARIO_PORT", "1886")

	if err := configParse(); err != nil {
		t.Errorf("unexpected error when parsing config: %v", err)
		return
	}

	if os.Getenv("PORT") != "1886" {
		t.Errorf("expected PORT=1886, but PORT=%s instead", os.Getenv("PORT"))
		return
	}
}

func TestConfigParseNoOrigin(t *testing.T) {
	_ = os.Setenv("COMENTARIO_ORIGIN", "")

	if err := configParse(); err == nil {
		t.Errorf("expected error not found parsing config without ORIGIN")
		return
	}
}

func TestConfigParseStatic(t *testing.T) {
	_ = os.Setenv("COMENTARIO_ORIGIN", "https://comentario.app")

	if err := configParse(); err != nil {
		t.Errorf("unexpected error when parsing config: %v", err)
		return
	}

	binPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		t.Errorf("cannot load binary path: %v", err)
		return
	}

	if os.Getenv("STATIC") != binPath {
		t.Errorf("COMENTARIO_STATIC != %s when unset", binPath)
		return
	}

	_ = os.Setenv("COMENTARIO_STATIC", "/usr/")

	if err := configParse(); err != nil {
		t.Errorf("unexpected error when parsing config: %v", err)
		return
	}

	if os.Getenv("STATIC") != "/usr" {
		t.Errorf("COMENTARIO_STATIC != /usr when unset")
		return
	}
}

func TestConfigParseStaticDNE(t *testing.T) {
	_ = os.Setenv("COMENTARIO_ORIGIN", "https://comentario.app")
	_ = os.Setenv("COMENTARIO_STATIC", "/does/not/exist/surely/")

	if err := configParse(); err == nil {
		t.Errorf("expected error not found when a non-existant directory is used")
		return
	}
}

func TestConfigParseStaticNotADirectory(t *testing.T) {
	_ = os.Setenv("COMENTARIO_ORIGIN", "https://comentario.app")
	_ = os.Setenv("COMENTARIO_STATIC", os.Args[0])

	if err := configParse(); err != errorNotADirectory {
		t.Errorf("expected error not found when a file is used")
		return
	}
}

func TestConfigOriginTrailingSlash(t *testing.T) {
	_ = os.Setenv("COMENTARIO_ORIGIN", "https://comentario.app/")
	_ = os.Setenv("COMENTARIO_STATIC", "")

	if err := configParse(); err != nil {
		t.Errorf("unexpected error when parsing config: %v", err)
		return
	}

	if os.Getenv("ORIGIN") != "https://comentario.app" {
		t.Errorf("expected ORIGIN=https://comentario.app got ORIGIN=%s", os.Getenv("ORIGIN"))
		return
	}
}

func TestConfigMaxConnections(t *testing.T) {
	_ = os.Setenv("COMENTARIO_ORIGIN", "https://comentario.app")
	_ = os.Setenv("COMENTARIO_STATIC", "")

	_ = os.Setenv("COMENTARIO_MAX_IDLE_PG_CONNECTIONS", "100")
	if err := configParse(); err != nil {
		t.Errorf("unexpected error when MAX_IDLE_PG_CONNECTIONS=100: %v", err)
		return
	}

	_ = os.Setenv("COMENTARIO_MAX_IDLE_PG_CONNECTIONS", "text")
	if err := configParse(); err == nil {
		t.Errorf("expected error with MAX_IDLE_PG_CONNECTIONS=text not found")
		return
	}

	_ = os.Setenv("COMENTARIO_MAX_IDLE_PG_CONNECTIONS", "0")
	if err := configParse(); err == nil {
		t.Errorf("expected error with MAX_IDLE_PG_CONNECTIONS=0 not found")
		return
	}

	_ = os.Setenv("COMENTARIO_MAX_IDLE_PG_CONNECTIONS", "-1")
	if err := configParse(); err == nil {
		t.Errorf("expected error with MAX_IDLE_PG_CONNECTIONS=-1 not found")
		return
	}
}
