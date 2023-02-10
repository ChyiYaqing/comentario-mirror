package api

import "gitlab.com/commento/commento/api/internal/util"

var commentersRowColumns = `
	commenters.commenterHex,
	commenters.email,
	commenters.name,
	commenters.link,
	commenters.photo,
	commenters.provider,
	commenters.joinDate
`

func commentersRowScan(s sqlScanner, c *commenter) error {
	return s.Scan(
		&c.CommenterHex,
		&c.Email,
		&c.Name,
		&c.Link,
		&c.Photo,
		&c.Provider,
		&c.JoinDate,
	)
}

func commenterGetByHex(commenterHex string) (commenter, error) {
	if commenterHex == "" {
		return commenter{}, util.ErrorMissingField
	}

	statement := `
		SELECT ` + commentersRowColumns + `
		FROM commenters
		WHERE commenterHex = $1;
	`
	row := DB.QueryRow(statement, commenterHex)

	var c commenter
	if err := commentersRowScan(row, &c); err != nil {
		// TODO: is this the only error?
		return commenter{}, util.ErrorNoSuchCommenter
	}

	return c, nil
}

func commenterGetByEmail(provider string, email string) (commenter, error) {
	if provider == "" || email == "" {
		return commenter{}, util.ErrorMissingField
	}

	statement := `
		SELECT ` + commentersRowColumns + `
		FROM commenters
		WHERE email = $1 AND provider = $2;
	`
	row := DB.QueryRow(statement, email, provider)

	var c commenter
	if err := commentersRowScan(row, &c); err != nil {
		// TODO: is this the only error?
		return commenter{}, util.ErrorNoSuchCommenter
	}

	return c, nil
}

func commenterGetByCommenterToken(commenterToken string) (commenter, error) {
	if commenterToken == "" {
		return commenter{}, util.ErrorMissingField
	}

	statement := `
		SELECT ` + commentersRowColumns + `
		FROM commenterSessions
		JOIN commenters ON commenterSessions.commenterHex = commenters.commenterHex
		WHERE commenterToken = $1;
	`
	row := DB.QueryRow(statement, commenterToken)

	var c commenter
	if err := commentersRowScan(row, &c); err != nil {
		// TODO: is this the only error?
		return commenter{}, util.ErrorNoSuchToken
	}

	if c.CommenterHex == "none" {
		return commenter{}, util.ErrorNoSuchToken
	}

	return c, nil
}
