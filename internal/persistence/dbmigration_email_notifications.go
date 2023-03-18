package persistence

import (
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/util"
	"time"
)

func migrateEmails(db *Database) error {
	rows, err := db.Query(`
		select commenters.email from commenters
		union
		select owners.email from owners
		union
		select moderators.email from moderators;
	`)
	if err != nil {
		logger.Errorf("cannot get comments: %v", err)
		return util.ErrorDatabaseMigration
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		if err = rows.Scan(&email); err != nil {
			logger.Errorf("cannot get email from tables during migration: %v", err)
			return util.ErrorDatabaseMigration
		}

		unsubscribeSecretHex, err := data.RandomHexID()
		if err != nil {
			return util.ErrorDatabaseMigration
		}

		err = db.Exec(
			"insert into emails(email, unsubscribesecrethex, lastemailnotificationdate) values ($1, $2, $3) "+
				"on conflict do nothing;",
			email,
			unsubscribeSecretHex,
			time.Now().UTC())
		if err != nil {
			logger.Errorf("cannot insert email during migration: %v", err)
			return util.ErrorDatabaseMigration
		}
	}

	return nil
}
