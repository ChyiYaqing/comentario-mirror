package api

import "gitlab.com/commento/commento/api/internal/util"

func commenterSessionUpdate(commenterToken string, commenterHex string) error {
	if commenterToken == "" || commenterHex == "" {
		return util.ErrorMissingField
	}

	statement := `update commenterSessions set commenterHex = $2 where commenterToken = $1;`
	_, err := DB.Exec(statement, commenterToken, commenterHex)
	if err != nil {
		logger.Errorf("error updating commenterHex: %v", err)
		return util.ErrorInternal
	}

	return nil
}
