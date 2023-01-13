package main

func domainOwnershipVerify(ownerHex string, domain string) (bool, error) {
	if ownerHex == "" || domain == "" {
		return false, errorMissingField
	}

	statement := `select EXISTS (select 1 from domains where ownerHex=$1 and domain=$2);`
	row := db.QueryRow(statement, ownerHex, domain)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		logger.Errorf("cannot query if domain owner: %v", err)
		return false, errorInternal
	}

	return exists, nil
}
