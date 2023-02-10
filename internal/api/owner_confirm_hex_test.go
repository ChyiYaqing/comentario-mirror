package api

import (
	"gitlab.com/commento/commento/api/internal/util"
	"testing"
	"time"
)

func TestOwnerConfirmHexBasics(t *testing.T) {
	FailTestOnError(t, SetupTestEnv())

	ownerHex, _ := ownerNew("test@example.com", "Test", "hunter2")

	statement := `
		update owners
		set confirmedEmail=false;
	`
	_, err := DB.Exec(statement)
	if err != nil {
		t.Errorf("unexpected error when setting confirmedEmail=false: %v", err)
		return
	}

	confirmHex, _ := util.RandomHex(32)

	statement = `
		insert into
		ownerConfirmHexes (confirmHex, ownerHex, sendDate)
		values            ($1,         $2,       $3      );
	`
	_, err = DB.Exec(statement, confirmHex, ownerHex, time.Now().UTC())
	if err != nil {
		t.Errorf("unexpected error creating inserting confirmHex: %v\n", err)
		return
	}

	if err = ownerConfirmHex(confirmHex); err != nil {
		t.Errorf("unexpected error confirming hex: %v", err)
		return
	}

	statement = `
		select confirmedEmail
		from owners
		where ownerHex=$1;
	`
	row := DB.QueryRow(statement, ownerHex)

	var confirmedHex bool
	if err = row.Scan(&confirmedHex); err != nil {
		t.Errorf("unexpected error scanning confirmedEmail: %v", err)
		return
	}

	if !confirmedHex {
		t.Errorf("confirmedHex expected to be true after confirmation; found to be false")
		return
	}
}
