package data

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"strings"
)

// EmailToString converts a value of *strfmt.Email into a string
func EmailToString(email *strfmt.Email) string {
	return TrimmedString((*string)(email))
}

// RandomHexID creates and returns a new, random hex ID
func RandomHexID() (models.HexID, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return models.HexID(hex.EncodeToString(b)), nil
}

// TrimmedString converts a *string value into a string, trimming all leading and trailing whitespace
func TrimmedString(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}
