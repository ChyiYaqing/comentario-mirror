package main

func commentDomainPathGet(commentHex string) (string, string, error) {
	if commentHex == "" {
		return "", "", errorMissingField
	}

	statement := `select domain, path from comments where commentHex = $1;`
	row := db.QueryRow(statement, commentHex)

	var domain string
	var path string
	var err error
	if err = row.Scan(&domain, &path); err != nil {
		return "", "", errorNoSuchDomain
	}

	return domain, path, nil
}
