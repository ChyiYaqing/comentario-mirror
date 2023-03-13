package data

import (
	"gitlab.com/comentario/comentario/internal/api/models"
	"strings"
)

// StringHexID converts a value of *models.HexID into a string
func StringHexID(id *models.HexID) string {
	if id == nil {
		return ""
	}
	return string(*id)
}

// TrimmedString converts a *string value into a string, trimming all leading and trailing whitespace
func TrimmedString(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}
