package svc

import (
	"database/sql"
	"errors"
	"github.com/op/go-logging"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/data"
)

// logger represents a package-wide logger instance
var logger = logging.MustGetLogger("svc")

var (
	ErrDB            = errors.New("services: database error")
	ErrNotFound      = errors.New("services: object not found")
	ErrUnknownEntity = errors.New("services: unknown entity")
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

// fixCommenterHex handles the anonymous commenter hex ID when persisting a database record.
func fixCommenterHex(id models.HexID) string {
	if id == data.AnonymousCommenter.HexID {
		return "anonymous"
	}
	return string(id)
}

// fixIdP handles default value (i.e. local authentication) for the identity provider when persisting a database record.
func fixIdP(idp string) string {
	// IdP defaults to local
	if idp == "" {
		return "commento"
	}
	return idp
}

// fixNone returns "none" if s is empty; meant for persisting a database record.
func fixNone(id models.HexID) string {
	if id == "" {
		return "none"
	}
	return string(id)
}

// fixUndefined returns "undefined" if s is empty; meant for persisting a database record.
func fixUndefined(s string) string {
	if s == "" {
		return "undefined"
	}
	return s
}

// translateDBErrors "translates" database errors into a service error, picking the first non-nil error
func translateDBErrors(errs ...error) error {
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

// unfixCommenterHex handles the anonymous commenter hex ID when reading a database record.
func unfixCommenterHex(id string) models.HexID {
	if id == "anonymous" {
		return data.AnonymousCommenter.HexID
	}
	return models.HexID(id)
}

// unfixIdP handles the default value (i.e. local authentication) for the identity provider when reading a database record.
func unfixIdP(idp string) string {
	if idp == "commento" {
		return ""
	}
	return idp
}

// unfixNone returns an empty string if s is "none"; meant for reading a database record.
func unfixNone(s string) models.HexID {
	if s == "none" {
		return ""
	}
	return models.HexID(s)
}

// unfixUndefined returns an empty string if s is "undefined"; meant for reading a database record.
func unfixUndefined(s string) string {
	if s == "undefined" {
		return ""
	}
	return s
}
