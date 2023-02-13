package persistence

import (
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

		unsubscribeSecretHex, err := util.RandomHex(32)
		if err != nil {
			return util.ErrorDatabaseMigration
		}

		statement := `insert into emails(email, unsubscribeSecretHex, lastEmailNotificationDate) values ($1,    $2,                   $3                       ) on conflict do nothing;`
		_, err = db.Exec(statement, email, unsubscribeSecretHex, time.Now().UTC())
		if err != nil {
			logger.Errorf("cannot insert email during migration: %v", err)
			return util.ErrorDatabaseMigration
		}
	}

	return nil
}
