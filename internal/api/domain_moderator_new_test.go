package api

import (
	"testing"
)

func TestDomainModeratorNewBasics(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	if err := domainModeratorNew("example.com", "test@example.com"); err != nil {
		t.Errorf("unexpected error creating new domain moderator: %v", err)
		return
	}
}

func TestDomainModeratorNewEmpty(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	if err := domainModeratorNew("example.com", ""); err == nil {
		t.Errorf("expected error not found when creating new moderator with empty email")
		return
	}

	if err := domainModeratorNew("", "test@example.com"); err == nil {
		t.Errorf("expected error not found when creating new moderator with empty domain")
		return
	}
}
