package config

import (
	"fmt"
	"github.com/op/go-logging"
	"gitlab.com/comentario/comentario/internal/util"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// KeySecret is a record containing a key and a secret
type KeySecret struct {
	Disable bool   `yaml:"disable"` // Can be used to forcefully disable the corresponding functionality
	Key     string `yaml:"key"`     // Public key
	Secret  string `yaml:"secret"`  // Private key
}

// Usable returns whether the instance isn't disabled and the key and the secret are filled in
func (c *KeySecret) Usable() bool {
	return !c.Disable && c.Key != "" && c.Secret != ""
}

var (
	AppVersion string // Application version set during bootstrapping
	BuildDate  string // Application build date set during bootstrapping
)

// logger represents a package-wide logger instance
var logger = logging.MustGetLogger("config")

var (
	// SecretsConfig is a configuration object for storing sensitive information
	SecretsConfig = &struct {
		Postgres struct {
			Host     string `yaml:"host"`     // PostgreSQL host
			Port     int    `yaml:"port"`     // PostgreSQL port
			Username string `yaml:"username"` // PostgreSQL username
			Password string `yaml:"password"` // PostgreSQL password
			Database string `yaml:"database"` // PostgreSQL database
			SSLMode  string `yaml:"sslmode"`  // PostgreSQL sslmode, defaults to "disable"
		} `yaml:"postgres"`

		SMTPServer struct {
			Host string `yaml:"host"`     // SMTP server hostname
			Port int    `yaml:"port"`     // SMTP server port
			User string `yaml:"username"` // SMTP server username
			Pass string `yaml:"password"` // SMTP server password
		} `yaml:"smtpServer"`

		IdP struct {
			GitHub  KeySecret `yaml:"github"`  // GitHub auth config
			GitLab  KeySecret `yaml:"gitlab"`  // GitLab auth config
			Google  KeySecret `yaml:"google"`  // Google auth config
			Twitter KeySecret `yaml:"twitter"` // Twitter auth config
		} `yaml:"idp"`

		Akismet struct {
			Key string `yaml:"key"` // Akismet key
		} `yaml:"akismet"`
	}{}

	// CLIFlags stores command-line flags
	CLIFlags = struct {
		Verbose         []bool `short:"v" long:"verbose" description:"Verbose logging"`
		BaseURL         string `long:"base-url"          description:"Server's own base URL"                      default:"http://localhost:8080/" env:"BASE_URL"`
		CDNURL          string `long:"cdn-url"           description:"Static file CDN URL (defaults to base URL)" default:""                       env:"CDN_URL"`
		EmailFrom       string `long:"email-from"        description:"'From' address in sent emails"              default:"noreply@localhost"      env:"EMAIL_FROM"`
		DBIdleConns     int    `long:"db-idle-conns"     description:"Max. # of idle DB connections"              default:"50"                     env:"DB_MAX_IDLE_CONNS"`
		EnableSwaggerUI bool   `long:"enable-swagger-ui" description:"Enable Swagger UI at /api/docs"`
		StaticPath      string `long:"static-path"       description:"Path to static files"                       default:"./frontend"             env:"STATIC_PATH"`
		DBMigrationPath string `long:"db-migration-path" description:"Path to DB migration files"                 default:"./db"                   env:"DB_MIGRATION_PATH"`
		TemplatePath    string `long:"template-path"     description:"Path to template files"                     default:"./templates"            env:"TEMPLATE_PATH"`
		SecretsFile     string `long:"secrets"           description:"Path to YAML file with secrets"             default:"secrets.yaml"           env:"SECRETS_FILE"`
		AllowNewOwners  bool   `long:"allow-new-owners"  description:"Allow new owner signups"                                                     env:"ALLOW_NEW_OWNERS"`
		GitLabURL       string `long:"gitlab-url"        description:"Custom GitLab URL for authentication"       default:""                       env:"GITLAB_URL"`
		E2e             bool   `long:"e2e"               description:"End-2-end testing mode"`
	}{}

	// Derived values

	BaseURL        *url.URL // The parsed base URL
	CDNURL         *url.URL // The parsed CDN URL
	UseHTTPS       bool     // Whether the base URL is a HTTPS one
	SMTPConfigured bool     // Whether sending emails is properly configured
)

// CLIParsed is a callback that signals the config the CLI flags have been parsed
func CLIParsed() error {
	// Parse the base URL
	var err error
	if BaseURL, err = util.ParseAbsoluteURL(CLIFlags.BaseURL); err != nil {
		return fmt.Errorf("invalid Base URL: %v", err)
	}
	UseHTTPS = BaseURL.Scheme == "https"

	// Check the CDN URL: if it's empty, use the base URL instead
	if CLIFlags.CDNURL == "" {
		CDNURL = BaseURL

	} else if CDNURL, err = util.ParseAbsoluteURL(CLIFlags.CDNURL); err != nil {
		return fmt.Errorf("invalid CDN URL: %v", err)
	}

	// Load secrets
	if err := UnmarshalConfigFile(CLIFlags.SecretsFile, SecretsConfig); err != nil {
		return err
	}

	// Configure OAuth providers
	oauthConfigure()

	// If SMTP credentials are available, use a corresponding mailer
	if SecretsConfig.SMTPServer.Host != "" && SecretsConfig.SMTPServer.User != "" && SecretsConfig.SMTPServer.Pass != "" {
		util.AppMailer = util.NewSMTPMailer(
			SecretsConfig.SMTPServer.Host,
			SecretsConfig.SMTPServer.Port,
			SecretsConfig.SMTPServer.User,
			SecretsConfig.SMTPServer.Pass,
			CLIFlags.EmailFrom)
		SMTPConfigured = true
	}

	// Succeeded
	return nil
}

// GuessUserLanguage tries to identify the most appropriate language for the user based on the request URL path, the
// user's language cookie and/or browser preferences, amongst those supported, and returns it as a 2-letter code.
func GuessUserLanguage(r *http.Request) string {
	// First, analyze the requested path. If it's under a language root, use that language
	if ok, p := PathOfBaseURL(r.URL.Path); ok && len(p) >= 3 && p[2] == '/' && util.IsUILang(p[0:2]) {
		return p[0:2]
	}

	// Next, try to extract the preferred language from a cookie
	cookieLang := ""
	if c, _ := r.Cookie("lang"); c != nil {
		cookieLang = c.Value
	}

	// Find the best match based on the cookie and/or browser header
	tag, _ := language.MatchStrings(util.UILangMatcher, cookieLang, r.Header.Get("Accept-Language"))
	base, _ := tag.Base()
	return base.String()
}

// PathOfBaseURL returns whether the given path is under the Base URL's path, and the path part relative to the base
// path (omitting the leading '/', if any)
func PathOfBaseURL(path string) (bool, string) {
	if strings.HasPrefix(path, BaseURL.Path) {
		return true, strings.TrimPrefix(path[len(BaseURL.Path):], "/")
	}
	return false, ""
}

// UnmarshalConfigFile reads in the specified YAML file at the specified path and unmarshalls it into the given variable
func UnmarshalConfigFile(filename string, out any) error {
	// Read in the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Unmarshal the data
	return yaml.Unmarshal(data, out)
}

// URLFor returns the complete absolute URL for the given path, with optional query params
func URLFor(path string, queryParams map[string]string) string {
	u := url.URL{
		Scheme: BaseURL.Scheme,
		Host:   BaseURL.Host,
		Path:   strings.TrimSuffix(BaseURL.Path, "/") + "/" + strings.TrimPrefix(path, "/"),
	}
	if queryParams != nil {
		q := url.Values{}
		for k, v := range queryParams {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// URLForAPI returns the complete absolute URL for the given API path, with optional query params
func URLForAPI(path string, queryParams map[string]string) string {
	return URLFor(util.APIPath+strings.TrimPrefix(path, "/"), queryParams)
}
