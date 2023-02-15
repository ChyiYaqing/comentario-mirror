package config

import (
	"gitlab.com/comentario/comentario/internal/util"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestPathOfBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		path     string
		wantOK   bool
		wantPath string
	}{
		{"domain root, empty        ", "http://api.test/", "", false, ""},
		{"domain root, root         ", "http://api.test/", "/", true, ""},
		{"subpath, empty            ", "http://api.test/some/path", "", false, ""},
		{"subpath, root             ", "http://api.test/some/path", "/", false, ""},
		{"subpath, same path, with /", "http://api.test/some/path", "/some/path", true, ""},
		{"subpath, same path, no /  ", "http://api.test/some/path", "some/path", false, ""},
		{"subpath, deep path, with /", "http://api.test/some/path", "/some/path/subpath", true, "subpath"},
		{"subpath, deep path, no /  ", "http://api.test/some/path", "some/path/subpath", false, ""},
	}
	for _, tt := range tests {
		var err error
		if BaseURL, err = url.Parse(tt.baseURL); err != nil {
			t.Errorf("PathOfBaseURL(): failed to parse base URL: %v", err)
		}
		t.Run(tt.name, func(t *testing.T) {
			gotOK, gotPath := PathOfBaseURL(tt.path)
			if gotOK != tt.wantOK {
				t.Errorf("PathOfBaseURL() got OK = %v, want %v", gotOK, tt.wantOK)
			}
			if gotPath != tt.wantPath {
				t.Errorf("PathOfBaseURL() got path = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

func TestURLFor(t *testing.T) {
	tests := []struct {
		name        string
		base        string
		path        string
		queryParams map[string]string
		want        string
	}{
		{"Root, no params", "http://ace.of.base:1234", "", nil, "http://ace.of.base:1234/"},
		{"Root with params", "http://basics/", "", map[string]string{"foo": "bar"}, "http://basics/?foo=bar"},
		{"Path, no params", "https://microsoft.qq:14/", "user/must/suffer", nil, "https://microsoft.qq:14/user/must/suffer"},
		{"Path with params", "https://yellow/submarine", "strawberry/fields", map[string]string{"baz": "   "}, "https://yellow/submarine/strawberry/fields?baz=+++"},
	}
	for _, tt := range tests {
		var err error
		t.Run(tt.name, func(t *testing.T) {
			BaseURL, err = url.Parse(tt.base)
			if err != nil {
				panic(err)
			}
			if got := URLFor(tt.path, tt.queryParams); got != tt.want {
				t.Errorf("URLFor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigParseBasics(t *testing.T) {
	_ = os.Setenv("COMENTARIO_ORIGIN", "https://comentario.app")

	if err := ConfigParse(); err != nil {
		t.Errorf("unexpected error when parsing config: %v", err)
		return
	}

	if os.Getenv("BIND_ADDRESS") != "127.0.0.1" {
		t.Errorf("expected COMENTARIO_BIND_ADDRESS=127.0.0.1, but COMENTARIO_BIND_ADDRESS=%s instead", os.Getenv("BIND_ADDRESS"))
		return
	}

	_ = os.Setenv("COMENTARIO_BIND_ADDRESS", "192.168.1.100")

	_ = os.Setenv("COMENTARIO_PORT", "")
	if err := ConfigParse(); err != nil {
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

	if err := ConfigParse(); err != nil {
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

	if err := ConfigParse(); err == nil {
		t.Errorf("expected error not found parsing config without ORIGIN")
		return
	}
}

func TestConfigParseStatic(t *testing.T) {
	_ = os.Setenv("COMENTARIO_ORIGIN", "https://comentario.app")

	if err := ConfigParse(); err != nil {
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

	if err := ConfigParse(); err != nil {
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

	if err := ConfigParse(); err == nil {
		t.Errorf("expected error not found when a non-existant directory is used")
		return
	}
}

func TestConfigParseStaticNotADirectory(t *testing.T) {
	_ = os.Setenv("COMENTARIO_ORIGIN", "https://comentario.app")
	_ = os.Setenv("COMENTARIO_STATIC", os.Args[0])

	if err := ConfigParse(); err != util.ErrorNotADirectory {
		t.Errorf("expected error not found when a file is used")
		return
	}
}

func TestConfigOriginTrailingSlash(t *testing.T) {
	_ = os.Setenv("COMENTARIO_ORIGIN", "https://comentario.app/")
	_ = os.Setenv("COMENTARIO_STATIC", "")

	if err := ConfigParse(); err != nil {
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
	if err := ConfigParse(); err != nil {
		t.Errorf("unexpected error when MAX_IDLE_PG_CONNECTIONS=100: %v", err)
		return
	}

	_ = os.Setenv("COMENTARIO_MAX_IDLE_PG_CONNECTIONS", "text")
	if err := ConfigParse(); err == nil {
		t.Errorf("expected error with MAX_IDLE_PG_CONNECTIONS=text not found")
		return
	}

	_ = os.Setenv("COMENTARIO_MAX_IDLE_PG_CONNECTIONS", "0")
	if err := ConfigParse(); err == nil {
		t.Errorf("expected error with MAX_IDLE_PG_CONNECTIONS=0 not found")
		return
	}

	_ = os.Setenv("COMENTARIO_MAX_IDLE_PG_CONNECTIONS", "-1")
	if err := ConfigParse(); err == nil {
		t.Errorf("expected error with MAX_IDLE_PG_CONNECTIONS=-1 not found")
		return
	}
}
