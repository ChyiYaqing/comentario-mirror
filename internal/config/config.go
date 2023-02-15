package config

import (
	"fmt"
	"github.com/op/go-logging"
	"gitlab.com/comentario/comentario/internal/util"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	AppVersion string // Application version set during bootstrapping
	BuildDate  string // Application build date set during bootstrapping
)

// logger represents a package-wide logger instance
var logger = logging.MustGetLogger("config")

var (

	// CLIFlags stores command-line flags
	CLIFlags = struct {
		Verbose          []bool `short:"v" long:"verbose"            description:"Verbose logging"`
		BaseURL          string `long:"base-url"                     description:"Server's own base URL"                             default:"http://localhost:8080/" env:"BASE_URL"`
		CDNURL           string `long:"cdn-url"                      description:"Static file CDN URL. If omitted, base URL is used" default:""                       env:"CDN_URL"`
		DBHost           string `long:"db-host"                      description:"PostgreSQL host"               default:"localhost"  env:"POSTGRES_HOST"`
		DBPort           int    `long:"db-port"                      description:"PostgreSQL port"               default:"5432"       env:"POSTGRES_PORT"`
		DBUsername       string `long:"db-username"                  description:"PostgreSQL username"           default:"postgres"   env:"POSTGRES_USERNAME"`
		DBPassword       string `long:"db-password"                  description:"PostgreSQL password"           default:"postgres"   env:"POSTGRES_PASSWORD"`
		DBName           string `long:"db-name"                      description:"PostgreSQL database name"      default:"comentario" env:"POSTGRES_DATABASE"`
		DBIdleConns      int    `long:"db-idle-conns"                description:"Max. # of idle DB connections" default:"50"         env:"DB_MAX_IDLE_CONNS"`
		DBMigrationsPath string `short:"m" long:"db-migrations-path" description:"Path to DB migration files"    default:"./db"       env:"DB_MIGRATIONS_PATH"`
		EnableSwaggerUI  bool   `long:"enable-swagger-ui"            description:"Enable Swagger UI at /api/docs"`
		StaticPath       string `short:"s" long:"static-path"        description:"Path to static files"          default:"."          env:"STATIC_PATH"`
		AllowNewOwners   bool   `long:"allow-new-owners"             description:"Allow new owner signups"                            env:"ALLOW_NEW_OWNERS"`
	}{}

	// Derived values

	BaseURL *url.URL // The parsed base URL
	CDNURL  *url.URL // The parsed CDN URL
)

// CLIParsed is a callback that signals the config the CLI flags have been parsed
func CLIParsed() error {
	// Parse the base URL
	var err error
	if BaseURL, err = util.ParseAbsoluteURL(CLIFlags.BaseURL); err != nil {
		return fmt.Errorf("invalid Base URL: %v", err)
	}

	// Check the CDN URL: if it's empty, use the base URL instead
	if CLIFlags.CDNURL == "" {
		CDNURL = BaseURL

	} else if CDNURL, err = util.ParseAbsoluteURL(CLIFlags.CDNURL); err != nil {
		return fmt.Errorf("invalid CDN URL: %v", err)
	}
	return nil
}

// PathOfBaseURL returns whether the given path is under the Base URL's path, and the path part relative to the base
// path (omitting the leading '/', if any)
func PathOfBaseURL(path string) (bool, string) {
	if strings.HasPrefix(path, BaseURL.Path) {
		return true, strings.TrimPrefix(path[len(BaseURL.Path):], "/")
	}
	return false, ""
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

// deprecated
func ConfigParse() error {
	binPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logger.Errorf("cannot load binary path: %v", err)
		return err
	}

	defaults := map[string]string{
		"CONFIG_FILE": "",

		"POSTGRES": "postgres://postgres:postgres@localhost/comentario?sslmode=disable",

		// PostgreSQL recommends max_connections in the order of hundreds. The default
		// is 100, so let's use half that and leave the other half for other services.
		// Ideally, you'd be setting this to a much higher number (for example, at the
		// time of writing, commento.io uses 600). See https://wiki.postgresql.org/wiki/Number_Of_Database_Connections
		"MAX_IDLE_PG_CONNECTIONS": "50",

		"BIND_ADDRESS": "127.0.0.1",
		"PORT":         "8080",
		"ORIGIN":       "",

		"CDN_PREFIX": "",

		"FORBID_NEW_OWNERS": "false",

		"STATIC": binPath,

		"GZIP_STATIC": "false",

		"SMTP_USERNAME":     "",
		"SMTP_PASSWORD":     "",
		"SMTP_HOST":         "",
		"SMTP_PORT":         "",
		"SMTP_FROM_ADDRESS": "",

		"AKISMET_KEY": "",

		"GOOGLE_KEY":    "",
		"GOOGLE_SECRET": "",

		"GITHUB_KEY":    "",
		"GITHUB_SECRET": "",

		"GITLAB_KEY":    "",
		"GITLAB_SECRET": "",
		"GITLAB_URL":    "https://gitlab.com",
	}

	if os.Getenv("COMENTARIO_CONFIG_FILE") != "" {
		if err := configFileLoad(os.Getenv("COMENTARIO_CONFIG_FILE")); err != nil {
			return err
		}
	}

	for key, value := range defaults {
		var err error
		if os.Getenv("COMENTARIO_"+key) == "" {
			err = os.Setenv(key, value)
		} else {
			err = os.Setenv(key, os.Getenv("COMENTARIO_"+key))
		}
		if err != nil {
			return err
		}
	}

	// Mandatory config parameters
	for _, env := range []string{"POSTGRES", "PORT", "ORIGIN", "FORBID_NEW_OWNERS", "MAX_IDLE_PG_CONNECTIONS"} {
		if os.Getenv(env) == "" {
			logger.Errorf("missing COMENTARIO_%s environment variable", env)
			return util.ErrorMissingConfig
		}
	}

	if err := os.Setenv("ORIGIN", strings.TrimSuffix(os.Getenv("ORIGIN"), "/")); err != nil {
		return err
	}
	if err := os.Setenv("ORIGIN", addHttpIfAbsent(os.Getenv("ORIGIN"))); err != nil {
		return err
	}

	if os.Getenv("CDN_PREFIX") == "" {
		if err := os.Setenv("CDN_PREFIX", os.Getenv("ORIGIN")); err != nil {
			return err
		}
	}

	if err := os.Setenv("CDN_PREFIX", strings.TrimSuffix(os.Getenv("CDN_PREFIX"), "/")); err != nil {
		return err
	}
	if err := os.Setenv("CDN_PREFIX", addHttpIfAbsent(os.Getenv("CDN_PREFIX"))); err != nil {
		return err
	}

	if os.Getenv("FORBID_NEW_OWNERS") != "true" && os.Getenv("FORBID_NEW_OWNERS") != "false" {
		logger.Errorf("COMENTARIO_FORBID_NEW_OWNERS neither 'true' nor 'false'")
		return util.ErrorInvalidConfigValue
	}

	static := os.Getenv("STATIC")
	for strings.HasSuffix(static, "/") {
		static = static[0 : len(static)-1]
	}

	file, err := os.Stat(static)
	if err != nil {
		logger.Errorf("cannot load %s: %v", static, err)
		return err
	}

	if !file.IsDir() {
		logger.Errorf("COMENTARIO_STATIC=%s is not a directory", static)
		return util.ErrorNotADirectory
	}

	if err := os.Setenv("STATIC", static); err != nil {
		return err
	}

	if num, err := strconv.Atoi(os.Getenv("MAX_IDLE_PG_CONNECTIONS")); err != nil {
		logger.Errorf("invalid COMENTARIO_MAX_IDLE_PG_CONNECTIONS: %v", err)
		return util.ErrorInvalidConfigValue
	} else if num <= 0 {
		logger.Errorf("COMENTARIO_MAX_IDLE_PG_CONNECTIONS should be a positive integer")
		return util.ErrorInvalidConfigValue
	}

	return nil
}

func addHttpIfAbsent(in string) string {
	if !strings.HasPrefix(in, "http://") && !strings.HasPrefix(in, "https://") {
		return "http://" + in
	}

	return in
}
