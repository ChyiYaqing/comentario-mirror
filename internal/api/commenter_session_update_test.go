package api

import (
	"testing"
)

func TestCommenterSessionUpdateBasics(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	commenterToken, _ := commenterTokenNew()

	if err := commenterSessionUpdate(commenterToken, "temp-commenter-hex"); err != nil {
		t.Errorf("unexpected error updating commenter session: %v", err)
		return
	}

	statement := `
		select commenterHex
		from commenterSessions
		where commenterToken = $1;
	`
	row := DB.QueryRow(statement, commenterToken)

	var commenterHex string
	if err := row.Scan(&commenterHex); err != nil {
		t.Errorf("error scanning commenterHex: %v", err)
		return
	}

	if commenterHex != "temp-commenter-hex" {
		t.Errorf("expected commenterHex=temp-commenter-hex got commenterHex=%s", commenterHex)
		return
	}
}

func TestCommenterSessionUpdateEmpty(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	if err := commenterSessionUpdate("", "temp-commenter-hex"); err == nil {
		t.Errorf("expected error not found when updating with empty commenterToken")
		return
	}
}
