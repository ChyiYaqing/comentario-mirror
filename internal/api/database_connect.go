package api

import (
	"database/sql"
	_ "github.com/lib/pq"
	"net/url"
	"os"
	"strconv"
	"time"
)

func DBConnect(retriesLeft int, retryDelay time.Duration) error {
	con := os.Getenv("POSTGRES")
	u, err := url.Parse(con)
	if err != nil {
		logger.Errorf("invalid postgres connection URI: %v", err)
		return err
	}
	u.User = url.UserPassword(u.User.Username(), "redacted")
	logger.Infof("opening connection to postgres: %s", u.String())

	DB, err = sql.Open("postgres", con)
	if err != nil {
		logger.Errorf("cannot open connection to postgres: %v", err)
		return err
	}

	err = DB.Ping()
	if err != nil {
		if retriesLeft > 0 {
			logger.Errorf("cannot talk to postgres, retrying in %s (%d attempts left): %v", retryDelay.String(), retriesLeft-1, err)
			time.Sleep(retryDelay)
			return DBConnect(retriesLeft-1, retryDelay*2)
		}
		logger.Errorf("cannot talk to postgres, last attempt failed: %v", err)
		return err
	}

	statement := `create table if not exists migrations (filename text not null unique);`
	_, err = DB.Exec(statement)
	if err != nil {
		logger.Errorf("cannot create migrations table: %v", err)
		return err
	}

	maxIdleConnections, err := strconv.Atoi(os.Getenv("MAX_IDLE_PG_CONNECTIONS"))
	if err != nil {
		logger.Warningf("cannot parse COMENTARIO_MAX_IDLE_PG_CONNECTIONS: %v", err)
		maxIdleConnections = 50
	}

	DB.SetMaxIdleConns(maxIdleConnections)

	return nil
}
