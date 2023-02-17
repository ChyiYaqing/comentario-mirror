package config

import (
	"net/url"
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
		{"Root, no params ", "http://ace.of.base:1234", "", nil, "http://ace.of.base:1234/"},
		{"Root with params", "http://basics/", "", map[string]string{"foo": "bar"}, "http://basics/?foo=bar"},
		{"Path, no params ", "https://microsoft.qq:14/", "user/must/suffer", nil, "https://microsoft.qq:14/user/must/suffer"},
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

func TestURLForAPI(t *testing.T) {
	tests := []struct {
		name        string
		base        string
		path        string
		queryParams map[string]string
		want        string
	}{
		{"Root, no params ", "http://ace.of.base:1234", "", nil, "http://ace.of.base:1234/api/"},
		{"Root with params", "http://basics/", "", map[string]string{"foo": "bar"}, "http://basics/api/?foo=bar"},
		{"Path, no params ", "https://microsoft.qq:14/", "user/must/suffer", nil, "https://microsoft.qq:14/api/user/must/suffer"},
		{"Path with params", "https://yellow/submarine", "strawberry/fields", map[string]string{"baz": "   "}, "https://yellow/submarine/api/strawberry/fields?baz=+++"},
	}
	for _, tt := range tests {
		var err error
		t.Run(tt.name, func(t *testing.T) {
			BaseURL, err = url.Parse(tt.base)
			if err != nil {
				panic(err)
			}
			if got := URLForAPI(tt.path, tt.queryParams); got != tt.want {
				t.Errorf("URLForAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}
