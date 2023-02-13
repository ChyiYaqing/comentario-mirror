package util

import (
	"github.com/op/go-logging"
	"regexp"
	"strings"
)

// logger represents a package-wide logger instance
var logger = logging.MustGetLogger("persistence")

var (
	reDNSHostname  = regexp.MustCompile(`^([a-z\d][-a-z\d]{0,62})(\.[a-z\d][-a-z\d]{0,62})*(\.[a-z]{2,63})$`) // Minimum 2 parts
	reEmailAddress = regexp.MustCompile(`^[^<>()[\]\\.,;:\s@"%]+(\.[^<>()[\]\\.,;:\s@"%]+)*@`)                // Only the part up to the '@'
)

// IsValidEmail returns whether the passed string is a valid email address
func IsValidEmail(s string) bool {
	// First validate the part before the '@'
	s = strings.ToLower(s)
	if s != "" && reEmailAddress.MatchString(s) {
		// Then the domain
		if i := strings.IndexByte(s, '@'); i > 0 {
			return IsValidHostname(s[i+1:])
		}
	}
	return false
}

// IsValidHostname returns true if the passed string is a valid domain hostname
func IsValidHostname(s string) bool {
	return s != "" && len(s) <= 253 && reDNSHostname.MatchString(s)
}
