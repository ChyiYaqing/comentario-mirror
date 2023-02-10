package api

import (
	"testing"
	"time"
)

func TestPageUpdateBasics(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "https://example.com/photo.jpg", "google", "")

	_, _ = commentNew(commenterHex, "example.com", "/path.html", "root", "**foo**", "unapproved", time.Now().UTC())

	p, _ := pageGet("example.com", "/path.html")
	if p.IsLocked != false {
		t.Errorf("expected IsLocked=false got %v", p.IsLocked)
		return
	}

	p.IsLocked = true

	if err := pageUpdate(p); err != nil {
		t.Errorf("unexpected error updating page: %v", err)
		return
	}

	p, _ = pageGet("example.com", "/path.html")
	if p.IsLocked != true {
		t.Errorf("expected IsLocked=true got %v", p.IsLocked)
		return
	}
}

func TestPageUpdateEmpty(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	p := page{Domain: "", Path: "", IsLocked: false}
	if err := pageUpdate(p); err == nil {
		t.Errorf("expected error not found updating page with empty everything")
		return
	}
}
