package util

import (
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/op/go-logging"
	"github.com/russross/blackfriday"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// logger represents a package-wide logger instance
var logger = logging.MustGetLogger("persistence")

var (
	reDNSHostname  = regexp.MustCompile(`^([a-z\d][-a-z\d]{0,62})(\.[a-z\d][-a-z\d]{0,62})*(\.[a-z]{2,63})$`) // Minimum 2 parts
	reEmailAddress = regexp.MustCompile(`^[^<>()[\]\\.,;:\s@"%]+(\.[^<>()[\]\\.,;:\s@"%]+)*@`)                // Only the part up to the '@'
)

// Scanner is a database/sql abstraction interface that can be used with both *sql.Row and *sql.Rows
type Scanner interface {
	// Scan copies columns from the underlying query row(s) to the values pointed to by dest
	Scan(dest ...any) error
}

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

// IsValidHostPort returns whether the passed string is a valid 'host' or 'host:port' spec, and its host and port values
func IsValidHostPort(s string) (bool, string, string) {
	// If there's a ':' in the string, we consider it the 'host:port' format (we ignore IPv6 colon-separated addresses
	// for now). Otherwise the entire string is considered a hostname
	host := s
	port := ""
	if i := strings.Index(s, ":"); i >= 0 {
		host = s[:i]
		// Validate port part
		port = s[i+1:]
		if port == "" || !IsValidPort(port) {
			return false, "", ""
		}
	}

	// Validate hostname, or otherwise  try to parse it as an IP address
	if IsValidHostname(host) || net.ParseIP(host) != nil {
		return true, host, port
	}
	return false, "", ""
}

// IsValidPort returns true if the passed string represents a valid port
func IsValidPort(s string) bool {
	i, err := strconv.Atoi(s)
	return err == nil && i > 0 && i < 65536
}

// MarkdownToHTML renders the provided markdown string as HTML
func MarkdownToHTML(markdown string) string {
	// Lazy-initialise the renderer
	if markdownRenderer == nil {
		createMarkdownRenderer()
	}

	// Render the markdown
	unsafe := blackfriday.Markdown([]byte(markdown), markdownRenderer, markdownExtensions)
	return string(markdownPolicy.SanitizeBytes(unsafe))
}

// ParseAbsoluteURL parses and returns the passed string as an absolute URL
func ParseAbsoluteURL(s string) (*url.URL, error) {
	// Parse the base URL
	var u *url.URL
	var err error
	if u, err = url.Parse(s); err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}

	// Check the scheme
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid URL scheme: '%s'", u.Scheme)
	}

	// Check the host
	if u.Host == "" {
		return nil, fmt.Errorf("invalid URL host: '%s'", u.Host)
	}

	// Verify it's a URL with a path starting with "/"
	if !strings.HasPrefix(u.Path, "/") {
		return nil, fmt.Errorf("invalid URL path (must begin with '/'): '%s'", u.Path)
	}

	// Remove any trailing slash from the base path, except when it's a root
	if len(u.Path) > 1 {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}
	return u, nil
}

// UserAgent return the value of the User-Agent request header
func UserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// UserIP tries to determine the user IP
func UserIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}

var markdownPolicy *bluemonday.Policy
var markdownRenderer blackfriday.Renderer

var markdownExtensions int

// createMarkdownRenderer creates and initialises a markdown renderer
func createMarkdownRenderer() {
	markdownPolicy = bluemonday.UGCPolicy()
	markdownPolicy.AddTargetBlankToFullyQualifiedLinks(true)
	markdownPolicy.RequireNoFollowOnFullyQualifiedLinks(true)

	markdownExtensions = 0
	markdownExtensions |= blackfriday.EXTENSION_AUTOLINK
	markdownExtensions |= blackfriday.EXTENSION_STRIKETHROUGH

	htmlFlags := 0
	htmlFlags |= blackfriday.HTML_SKIP_HTML
	htmlFlags |= blackfriday.HTML_SKIP_IMAGES
	htmlFlags |= blackfriday.HTML_SAFELINK
	htmlFlags |= blackfriday.HTML_HREF_TARGET_BLANK

	markdownRenderer = blackfriday.HtmlRenderer(htmlFlags, "", "")
}
