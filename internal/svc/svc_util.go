package svc

import (
	"database/sql"
	"errors"
	"github.com/op/go-logging"
)

// logger represents a package-wide logger instance
var logger = logging.MustGetLogger("svc")

var (
	ErrDB             = errors.New("services: database error")
	ErrDuplicateEmail = errors.New("services: duplicate email")
	ErrNotFound       = errors.New("services: object not found")
	ErrPageLocked     = errors.New("services: page is locked")
	ErrUnknownEntity  = errors.New("services: unknown entity")
)

// checkErrors picks and returns the first non-nil error, or nil if there's none
func checkErrors(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// translateErrors "translates" database errors into a service error, picking the first non-nil error
func translateErrors(errs ...error) error {
	switch checkErrors(errs...) {
	case nil:
		// No error
		return nil
	case sql.ErrNoRows:
		// Not found
		return ErrNotFound
	default:
		// Any other database error
		return ErrDB
	}
}
