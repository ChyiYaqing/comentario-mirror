package api

import (
	"gitlab.com/commento/commento/api/internal/util"
	"net/http"
)

func ownerDelete(ownerHex string, deleteDomains bool) error {
	domains, err := domainList(ownerHex)
	if err != nil {
		return err
	}

	if len(domains) > 0 {
		if !deleteDomains {
			return util.ErrorCannotDeleteOwnerWithActiveDomains
		}
		for _, d := range domains {
			if err := domainDelete(d.Domain); err != nil {
				return err
			}
		}
	}

	statement := `delete from owners where ownerHex = $1;`
	_, err = DB.Exec(statement, ownerHex)
	if err != nil {
		return util.ErrorNoSuchOwner
	}

	statement = `delete from ownersessions where ownerHex = $1;`
	_, err = DB.Exec(statement, ownerHex)
	if err != nil {
		logger.Errorf("cannot delete from ownersessions: %v", err)
		return util.ErrorInternal
	}

	statement = `delete from resethexes where hex = $1;`
	_, err = DB.Exec(statement, ownerHex)
	if err != nil {
		logger.Errorf("cannot delete from resethexes: %v", err)
		return util.ErrorInternal
	}

	return nil
}

func ownerDeleteHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
	}

	var x request
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	o, err := ownerGetByOwnerToken(*x.OwnerToken)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if err = ownerDelete(o.OwnerHex, false); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true})
}
