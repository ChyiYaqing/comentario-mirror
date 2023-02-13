package api

import (
	"fmt"
	"gitlab.com/comentario/comentario/internal/util"
	"os"
	"testing"
	"time"

	"github.com/op/go-logging"
)

func FailTestOnError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("failed test: %v", err)
	}
}

func getPublicTables() ([]string, error) {
	statement := `
		select tablename
		from pg_tables
		where schemaname='public';
	`
	rows, err := DB.Query(statement)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cannot query public tables: %v", err)
		return []string{}, err
	}

	defer rows.Close()

	tables := []string{}
	for rows.Next() {
		var table string
		if err = rows.Scan(&table); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cannot scan table name: %v", err)
			return []string{}, err
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func dropTables() error {
	tables, err := getPublicTables()
	if err != nil {
		return err
	}

	for _, table := range tables {
		if table != "migrations" {
			_, err = DB.Exec(fmt.Sprintf("drop table %s;", table))
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "cannot drop %s: %v", table, err)
				return err
			}
		}
	}

	return nil
}

func setupTestDatabase() error {
	if os.Getenv("COMENTARIO_POSTGRES") != "" {
		// set it manually because we need to use comentario_test, not comentario, by mistake
		_ = os.Setenv("POSTGRES", os.Getenv("COMENTARIO_POSTGRES"))
	} else {
		_ = os.Setenv("POSTGRES", "postgres://postgres:postgres@localhost/comentario_test?sslmode=disable")
	}

	if err := DBConnect(0, time.Second); err != nil {
		return err
	}

	if err := dropTables(); err != nil {
		return err
	}

	if err := MigrateFromDir("../../db/"); err != nil {
		return err
	}

	return nil
}

func clearTables() error {
	tables, err := getPublicTables()
	if err != nil {
		return err
	}

	for _, table := range tables {
		_, err = DB.Exec(fmt.Sprintf("delete from %s;", table))
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cannot clear %s: %v", table, err)
			return err
		}
	}

	return nil
}

var setupComplete bool

func SetupTestEnv() error {
	if !setupComplete {
		setupComplete = true

		if err := util.LoggerCreate(); err != nil {
			return err
		}

		// Print messages to console only if verbose. Sounds like a good idea to
		// keep the console clean on `go test`.
		if !testing.Verbose() {
			logging.SetLevel(logging.CRITICAL, "")
		}

		if err := setupTestDatabase(); err != nil {
			return err
		}

		if err := MarkdownRendererCreate(); err != nil {
			return err
		}
	}

	if err := clearTables(); err != nil {
		return err
	}

	return nil
}
