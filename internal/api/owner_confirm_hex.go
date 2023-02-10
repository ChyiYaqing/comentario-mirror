package api

import (
	"fmt"
	"gitlab.com/commento/commento/api/internal/util"
	"net/http"
	"os"
)

func ownerConfirmHex(confirmHex string) error {
	if confirmHex == "" {
		return util.ErrorMissingField
	}

	statement := `
		update owners
		set confirmedEmail=true
		where ownerHex in (
			select ownerHex from ownerConfirmHexes
			where confirmHex=$1
		);
	`
	res, err := DB.Exec(statement, confirmHex)
	if err != nil {
		logger.Errorf("cannot mark user's confirmedEmail as true: %v\n", err)
		return util.ErrorInternal
	}

	count, err := res.RowsAffected()
	if err != nil {
		logger.Errorf("cannot count rows affected: %v\n", err)
		return util.ErrorInternal
	}

	if count == 0 {
		return util.ErrorNoSuchConfirmationToken
	}

	statement = `
		delete from ownerConfirmHexes
		where confirmHex=$1;
	`
	_, err = DB.Exec(statement, confirmHex)
	if err != nil {
		logger.Warningf("cannot remove confirmation token: %v\n", err)
		// Don't return an error because this is not critical.
	}

	return nil
}

func ownerConfirmHexHandler(w http.ResponseWriter, r *http.Request) {
	if confirmHex := r.FormValue("confirmHex"); confirmHex != "" {
		if err := ownerConfirmHex(confirmHex); err == nil {
			http.Redirect(w, r, fmt.Sprintf("%s/login?confirmed=true", os.Getenv("ORIGIN")), http.StatusTemporaryRedirect)
			return
		}
	}

	// TODO: include error message in the URL
	http.Redirect(w, r, fmt.Sprintf("%s/login?confirmed=false", os.Getenv("ORIGIN")), http.StatusTemporaryRedirect)
}
