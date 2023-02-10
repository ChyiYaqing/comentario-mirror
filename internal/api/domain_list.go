package api

import (
	"gitlab.com/commento/commento/api/internal/util"
	"net/http"
)

func domainList(ownerHex string) ([]domain, error) {
	if ownerHex == "" {
		return []domain{}, util.ErrorMissingField
	}

	statement := `
		SELECT ` + domainsRowColumns + `
		FROM domains
		WHERE ownerHex=$1;
	`
	rows, err := DB.Query(statement, ownerHex)
	if err != nil {
		logger.Errorf("cannot query domains: %v", err)
		return nil, util.ErrorInternal
	}
	defer rows.Close()

	domains := []domain{}
	for rows.Next() {
		var d domain
		if err = domainsRowScan(rows, &d); err != nil {
			logger.Errorf("cannot Scan domain: %v", err)
			return nil, util.ErrorInternal
		}

		d.Moderators, err = domainModeratorList(d.Domain)
		if err != nil {
			return []domain{}, err
		}

		domains = append(domains, d)
	}

	return domains, rows.Err()
}

func domainListHandler(w http.ResponseWriter, r *http.Request) {
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

	domains, err := domainList(o.OwnerHex)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(
		w,
		response{
			"success": true,
			"domains": domains,
			"configuredOauths": map[string]bool{
				"google": googleConfigured,
				"github": githubConfigured,
				"gitlab": gitlabConfigured,
			},
		})
}
