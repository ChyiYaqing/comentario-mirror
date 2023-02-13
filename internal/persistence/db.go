package persistence

import (
	"database/sql"
	"fmt"
	"github.com/op/go-logging"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/util"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

// logger represents a package-wide logger instance
var logger = logging.MustGetLogger("persistence")

var goMigrations = map[string]func(db *Database) error{
	"20190213033530-email-notifications.sql": migrateEmails,
}

// Database is an opaque structure providing database operations
type Database struct {
	db *sql.DB // Internal SQL database instance
}

// InitDB establishes a database connection
func InitDB() (*Database, error) {
	db := &Database{}

	// Try to connect
	if err := db.connect(); err != nil {
		return nil, err
	}

	// Run migrations
	if err := db.migrate(); err != nil {
		return nil, err
	}

	// Succeeded
	return db, nil
}

// Exec executes the provided statement against the database
func (db *Database) Exec(query string, args ...any) (sql.Result, error) {
	return db.db.Exec(query, args...)
}

// Query executes the provided query against the database
func (db *Database) Query(query string, args ...any) (*sql.Rows, error) {
	return db.db.Query(query, args...)
}

// QueryRow queries a signle row from the database
func (db *Database) QueryRow(query string, args ...any) *sql.Row {
	return db.db.QueryRow(query, args...)
}

// Shutdown ends the database connection and shuts down all dependent services
func (db *Database) Shutdown() error {
	// If there's a connection, try to disconnect
	if db != nil {
		logger.Info("Disconnecting from database...")
		if err := db.db.Close(); err != nil {
			logger.Errorf("Failed to disconnect from database: %v", err)
		}
	}

	// Succeeded
	logger.Info("Disconnected from database")
	return nil
}

// connect establishes a database connection up to the configured number of attempts
func (db *Database) connect() error {
	logger.Infof("Connecting to database at %s@%s:%d...", config.CLIFlags.DBUsername, config.CLIFlags.DBHost, config.CLIFlags.DBPort)

	var err error
	var retryDelay = time.Second // Start with a delay of one second
	for attempt := 1; attempt <= util.DBMaxAttempts; attempt++ {
		// Try to establish a connection
		if err = db.tryConnect(attempt, util.DBMaxAttempts); err == nil {
			break // Succeeded
		}

		// Failed to connect. Wait a progressively doubling period of time before the next attempt
		time.Sleep(retryDelay)
		retryDelay *= 2
	}

	// Failed to connect
	if err != nil {
		logger.Errorf("Failed to connect to database after %d attempts, exiting", util.DBMaxAttempts)
		return err
	}

	// Configure the database
	db.db.SetMaxIdleConns(config.CLIFlags.DBIdleConns)
	return nil
}

// getAvailableMigrations returns a list of available database migration files
func (db *Database) getAvailableMigrations() ([]string, error) {
	// Scan the migrations dir for available migration files
	files, err := os.ReadDir(config.CLIFlags.DBMigrationsPath)
	if err != nil {
		logger.Errorf("Failed to read DB migrations dir '%s': %v", config.CLIFlags.DBMigrationsPath, err)
		return nil, err
	}

	// Convert the list of entries into a list of file names
	var list []string
	for _, file := range files {
		// Ignore directories and non-SQL files
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			list = append(list, file.Name())
		}
	}

	// The files must be sorted by name, in the ascending order
	sort.Strings(list)
	logger.Infof("Discovered %d database migrations in %s", len(list), config.CLIFlags.DBMigrationsPath)
	return list, err
}

// getInstalledMigrations returns a map of installed database migrations
func (db *Database) getInstalledMigrations() (map[string]bool, error) {
	// Query the migrations table
	rows, err := db.db.Query("select filename from migrations;")
	if err != nil {
		logger.Errorf("getInstalledMigrations: Query() failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	// Convert the files into a map
	m := make(map[string]bool)
	for rows.Next() {
		var s string
		if err = rows.Scan(&s); err != nil {
			logger.Errorf("getInstalledMigrations: Scan() failed: %v", err)
			return nil, err
		}
		m[s] = true
	}
	return m, nil
}

// migrate runs all known database migrations
func (db *Database) migrate() error {
	// Make sure the migrations table exists
	_, err := db.db.Exec(`create table if not exists migrations (filename text not null unique);`)
	if err != nil {
		logger.Errorf("Failed to create table 'migrations': %v", err)
		return err
	}

	// Read available migrations
	available, err := db.getAvailableMigrations()

	// Query already installed migrations
	installed, err := db.getInstalledMigrations()
	if err != nil {
		return err
	}
	logger.Infof("%d migrations already installed", len(installed))

	cntOK := 0
	for _, filename := range available {
		// Skip migrations that are already installed
		if installed[filename] {
			logger.Debugf("Migration '%s' is already installed", filename)
			continue
		}

		// Read in the content of the file
		logger.Debugf("Installing migration '%s'", filename)
		fullName := path.Join(config.CLIFlags.DBMigrationsPath, filename)
		contents, err := os.ReadFile(fullName)
		if err != nil {
			logger.Errorf("Failed to read file '%s': %v", fullName, err)
			return err
		}

		// Run the content of the file
		if _, err = db.db.Exec(string(contents)); err != nil {
			logger.Errorf("Failed to execute SQL in '%s': %v", fullName, err)
			return err
		}

		// Register the migration in the database
		if _, err = db.db.Exec("insert into migrations (filename) values ($1);", filename); err != nil {
			logger.Errorf("Failed to register migration '%s': %v", filename, err)
			return err
		}

		// Run the necessary code migrations
		if fn, ok := goMigrations[filename]; ok {
			if err = fn(db); err != nil {
				logger.Errorf("Failed to execute Go migration for '%s': %v", fullName, err)
				return err
			}
		}

		cntOK++
	}

	if cntOK > 0 {
		logger.Infof("Successfully installed %d new migrations", cntOK)
	} else {
		logger.Infof("No new migrations found")
	}
	return nil
}

// tryConnect tries to establish a database connection, once
func (db *Database) tryConnect(num, total int) error {
	var err error
	db.db, err = sql.Open(
		"postgres",
		fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			config.CLIFlags.DBUsername,
			config.CLIFlags.DBPassword,
			config.CLIFlags.DBHost,
			config.CLIFlags.DBPort,
			config.CLIFlags.DBName,
		))

	// Failed to connect
	if err != nil {
		logger.Warningf("[Attempt %d/%d] Failed to connect to database: %v", num, total, err)
		return err
	}

	// Connected successfully. Verify the connection by issuing a ping
	err = db.db.Ping()
	if err != nil {
		logger.Warningf("[Attempt %d/%d] Failed to ping database: %v", num, total, err)
	}
	return err
}
