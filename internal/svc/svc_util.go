package svc

import (
	"database/sql"
	"errors"
	"github.com/op/go-logging"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/util"
)

// logger represents a package-wide logger instance
var logger = logging.MustGetLogger("svc")

var (
	ErrNotFound     = errors.New("services: object not found")
	ErrInvalidInput = errors.New("services: invalid input passed")
	ErrDB           = errors.New("services: database error")
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

// validateEmail validates the passed email value and returns an ErrInvalidInput should it prove invalid
func validateEmail(s string) error {
	if !util.IsValidEmail(s) {
		return ErrInvalidInput
	}
	return nil
}

// validateHexID validates the passed hex ID value and returns an ErrInvalidInput should it prove invalid
func validateHexID(s models.HexID) error {
	if len(s) != 64 {
		return ErrInvalidInput
	}
	return nil
}
