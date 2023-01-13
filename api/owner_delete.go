package main

import (
	"net/http"
)

func ownerDelete(ownerHex string, deleteDomains bool) error {
	domains, err := domainList(ownerHex)
	if err != nil {
		return err
	}

	if len(domains) > 0 {
		if !deleteDomains {
			return errorCannotDeleteOwnerWithActiveDomains
		}
		for _, d := range domains {
			if err := domainDelete(d.Domain); err != nil {
				return err
			}
		}
	}

	statement := `delete from owners where ownerHex = $1;`
	_, err = db.Exec(statement, ownerHex)
	if err != nil {
		return errorNoSuchOwner
	}

	statement = `delete from ownersessions where ownerHex = $1;`
	_, err = db.Exec(statement, ownerHex)
	if err != nil {
		logger.Errorf("cannot delete from ownersessions: %v", err)
		return errorInternal
	}

	statement = `delete from resethexes where hex = $1;`
	_, err = db.Exec(statement, ownerHex)
	if err != nil {
		logger.Errorf("cannot delete from resethexes: %v", err)
		return errorInternal
	}

	return nil
}

func ownerDeleteHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
	}

	var x request
	if err := bodyUnmarshal(r, &x); err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	o, err := ownerGetByOwnerToken(*x.OwnerToken)
	if err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if err = ownerDelete(o.OwnerHex, false); err != nil {
		bodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	bodyMarshalChecked(w, response{"success": true})
}
