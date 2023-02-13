package config

import (
	"github.com/op/go-logging"
	"gitlab.com/comentario/comentario/internal/util"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var logger = logging.MustGetLogger("config")

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
