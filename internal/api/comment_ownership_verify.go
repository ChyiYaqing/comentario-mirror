package api

import "gitlab.com/commento/commento/api/internal/util"

func commentOwnershipVerify(commenterHex string, commentHex string) (bool, error) {
	if commenterHex == "" || commentHex == "" {
		return false, util.ErrorMissingField
	}

	statement := `select EXISTS(select 1 from comments where commenterHex=$1 and commentHex=$2);`
	row := DB.QueryRow(statement, commenterHex, commentHex)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		logger.Errorf("cannot query if comment owner: %v", err)
		return false, util.ErrorInternal
	}

	return exists, nil
}
