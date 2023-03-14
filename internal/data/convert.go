package data

import (
	"github.com/go-openapi/strfmt"
	"strings"
)

// EmailToString converts a value of *strfmt.Email into a string
func EmailToString(email *strfmt.Email) string {
	return TrimmedString((*string)(email))
}

// TrimmedString converts a *string value into a string, trimming all leading and trailing whitespace
func TrimmedString(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}
