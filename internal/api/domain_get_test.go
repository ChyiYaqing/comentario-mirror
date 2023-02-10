package api

import (
	"testing"
)

func TestDomainGetBasics(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	_ = domainNew("temp-owner-hex", "Example", "example.com")

	d, err := domainGet("example.com")
	if err != nil {
		t.Errorf("unexpected error getting domain: %v", err)
		return
	}

	if d.Name != "Example" {
		t.Errorf("expected name=Example got name=%s", d.Name)
		return
	}
}

func TestDomainGetEmpty(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	if _, err := domainGet(""); err == nil {
		t.Errorf("expected error not found when getting with empty domain")
		return
	}
}

func TestDomainGetDNE(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	if _, err := domainGet("example.com"); err == nil {
		t.Errorf("expected error not found when getting non-existant domain")
		return
	}
}
