package api

import "gitlab.com/commento/commento/api/internal/util"

func domainOwnershipVerify(ownerHex string, domain string) (bool, error) {
	if ownerHex == "" || domain == "" {
		return false, util.ErrorMissingField
	}

	statement := `select EXISTS (select 1 from domains where ownerHex=$1 and domain=$2);`
	row := DB.QueryRow(statement, ownerHex, domain)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		logger.Errorf("cannot query if domain owner: %v", err)
		return false, util.ErrorInternal
	}

	return exists, nil
}
