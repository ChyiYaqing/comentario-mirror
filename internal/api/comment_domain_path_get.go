package api

import (
	"gitlab.com/commento/commento/api/internal/util"
)

func commentDomainPathGet(commentHex string) (string, string, error) {
	if commentHex == "" {
		return "", "", util.ErrorMissingField
	}

	statement := `select domain, path from comments where commentHex = $1;`
	row := DB.QueryRow(statement, commentHex)

	var domain string
	var path string
	var err error
	if err = row.Scan(&domain, &path); err != nil {
		return "", "", util.ErrorNoSuchDomain
	}

	return domain, path, nil
}
