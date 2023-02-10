package api

import (
	"os"
	"strings"
)

var goMigrations = map[string]func() error{
	"20190213033530-email-notifications.sql": migrateEmails,
}

func Migrate() error {
	return MigrateFromDir(os.Getenv("STATIC") + "/db")
}

func MigrateFromDir(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		logger.Errorf("cannot read directory for migrations: %v", err)
		return err
	}

	statement := `select filename from migrations;`
	rows, err := DB.Query(statement)
	if err != nil {
		logger.Errorf("cannot query migrations: %v", err)
		return err
	}

	defer rows.Close()

	filenames := make(map[string]bool)
	for rows.Next() {
		var filename string
		if err = rows.Scan(&filename); err != nil {
			logger.Errorf("cannot scan filename: %v", err)
			return err
		}

		filenames[filename] = true
	}

	logger.Infof("%d migrations already installed, looking for more", len(filenames))

	completed := 0
	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, ".sql") {
			if !filenames[name] {
				f := dir + string(os.PathSeparator) + name
				contents, err := os.ReadFile(f)
				if err != nil {
					logger.Errorf("cannot read file %s: %v", name, err)
					return err
				}

				if _, err = DB.Exec(string(contents)); err != nil {
					logger.Errorf("cannot execute the SQL in %s: %v", f, err)
					return err
				}

				statement = ` insert into migrations (filename) values ($1); `
				_, err = DB.Exec(statement, name)
				if err != nil {
					logger.Errorf("cannot insert filename into the migrations table: %v", err)
					return err
				}

				if fn, ok := goMigrations[name]; ok {
					if err = fn(); err != nil {
						logger.Errorf("cannot execute Go migration associated with SQL %s: %v", f, err)
						return err
					}
				}

				completed++
			}
		}
	}

	if completed > 0 {
		logger.Infof("%d new migrations completed (%d total)", completed, len(filenames)+completed)
	} else {
		logger.Infof("none found")
	}

	return nil
}
