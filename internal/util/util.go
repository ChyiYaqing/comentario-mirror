package util

import (
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/op/go-logging"
	"github.com/russross/blackfriday"
	"golang.org/x/net/html"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Scanner is a database/sql abstraction interface that can be used with both *sql.Row and *sql.Rows
type Scanner interface {
	// Scan copies columns from the underlying query row(s) to the values pointed to by dest
	Scan(dest ...any) error
}

// Mailer allows sending emails
type Mailer interface {
	// Mail sends an email to the specified recipient.
	// replyTo:     email address/name of the sender (optional).
	// recipient:   email address/name of the recipient.
	// subject:     email subject.
	// htmlMessage: email text in the HTML format.
	Mail(replyTo, recipient, subject, htmlMessage string) error
}

// logger represents a package-wide logger instance
var logger = logging.MustGetLogger("persistence")

var (
	reHexID        = regexp.MustCompile(`^[\da-f]{64}$`)
	reDNSHostname  = regexp.MustCompile(`^([a-z\d][-a-z\d]{0,62})(\.[a-z\d][-a-z\d]{0,62})*$`) // Minimum one part
	reEmailAddress = regexp.MustCompile(`^[^<>()[\]\\.,;:\s@"%]+(\.[^<>()[\]\\.,;:\s@"%]+)*@`) // Only the part up to the '@'

	// AppMailer is a Mailer implementation available application-wide. Defaults to a mailer that doesn't do anything
	AppMailer Mailer = &noOpMailer{}
)

// ----------------------------------------------------------------------------------------------------------------------

// noOpMailer is a Mailer implementation that doesn't send any emails
type noOpMailer struct{}

func (m *noOpMailer) Mail(_, recipient, subject, _ string) error {
	logger.Debugf("NoOpMailer: not sending email to %s (subject: '%s')", recipient, subject)
	return nil
}

// ----------------------------------------------------------------------------------------------------------------------

// DownloadGzip downloads a gzip-compressed archive from the given URL, then decompresses it and returns the binary data
func DownloadGzip(dataURL string) ([]byte, error) {
	// Fetch the archive
	resp, err := http.Get(dataURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and decompress the data
	if r, err := gzip.NewReader(resp.Body); err != nil {
		return nil, err
	} else if b, err := io.ReadAll(r); err != nil {
		return nil, err
	} else {
		return b, nil
	}
}

// HTMLDocumentTitle parses and returns the title of an HTML document
func HTMLDocumentTitle(body io.Reader) (string, error) {
	// Iterate the body's tokens
	tokenizer := html.NewTokenizer(body)
	for {
		// Get the next token type
		switch tokenizer.Next() {
		// An error token, we either reached the end of the file, or the HTML was malformed
		case html.ErrorToken:
			err := tokenizer.Err()
			// End of the stream
			if err == io.EOF {
				return "", errors.New("no title found in HTML document")
			}

			// Any other error
			return "", tokenizer.Err()

		// A start tag token
		case html.StartTagToken:
			token := tokenizer.Token()
			// If it's the title tag
			if token.Data == "title" {
				// The next token should be the page title
				if tokenizer.Next() == html.TextToken {
					return tokenizer.Token().Data, nil
				}
			}
		}
	}
}

// HTMLTitleFromURL tries to fetch the specified URL and subsequently extract the title from its HTML document
func HTMLTitleFromURL(url string) (string, error) {
	// Fetch the URL
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Verify we're dealing with a HTML document
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		return "", nil
	}

	// Parse the response body
	return HTMLDocumentTitle(resp.Body)
}

// IsValidURL returns whether the passed string is a valid absolute URL
func IsValidURL(s string) bool {
	_, err := ParseAbsoluteURL(s)
	return err == nil
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

// IsValidHexID returns true if the passed string is a valid hex ID
func IsValidHexID(s string) bool {
	return len(s) == 64 && reHexID.MatchString(s)
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

// IsUILang returns whether the provided 2-letter string is a supported UI language
func IsUILang(s string) bool {
	// Only 2-letter codes are in scope
	if len(s) == 2 {
		// Search through the available languages to find one whose base matches the string
		for _, t := range UILanguageTags {
			if base, _ := t.Base(); base.String() == s {
				return true
			}
		}
	}
	return false
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

// RandomSleep sleeps a random duration of time within the given interval
func RandomSleep(min, max time.Duration) {
	time.Sleep(time.Duration(int64(min) + rand.Int63n(int64(max-min))))
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

// ---------------------------------------------------------------------------------------------------------------------

// SafeStringMap is a thread-safe map[string]string. Its zero value is a usable map
type SafeStringMap[K ~string] struct {
	m  map[K]string
	mu sync.Mutex
}

// Put stores a value under the given key
func (m *SafeStringMap[K]) Put(k K, v string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.m == nil {
		m.m = make(map[K]string)
	}
	m.m[k] = v
}

// Take retrieves and deletes a values by its key, thread-safely
func (m *SafeStringMap[K]) Take(k K) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Don't bother if there's no data
	if m.m == nil {
		return "", false
	}

	// Fetch the value
	v, ok := m.m[k]

	// Remove the entry
	delete(m.m, k)
	return v, ok
}

// Len returns the number of entries in the map
func (m *SafeStringMap[K]) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.m == nil {
		return 0
	}
	return len(m.m)
}
