package main

func commenterSessionUpdate(commenterToken string, commenterHex string) error {
	if commenterToken == "" || commenterHex == "" {
		return errorMissingField
	}

	statement := `update commenterSessions set commenterHex = $2 where commenterToken = $1;`
	_, err := db.Exec(statement, commenterToken, commenterHex)
	if err != nil {
		logger.Errorf("error updating commenterHex: %v", err)
		return errorInternal
	}

	return nil
}
