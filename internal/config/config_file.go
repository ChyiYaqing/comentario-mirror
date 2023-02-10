package config

import (
	"bufio"
	"gitlab.com/commento/commento/api/internal/util"
	"os"
	"strings"
)

func configFileLoad(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}

	defer file.Close()

	num := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		num += 1

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		i := strings.Index(line, "=")
		if i == -1 {
			logger.Errorf("%s: line %d: neither a comment nor a valid setting", filepath, num)
			return util.ErrorInvalidConfigFile
		}

		key := line[:i]
		value := line[i+1:]

		if !strings.HasPrefix(key, "COMENTARIO_") {
			continue
		}

		if os.Getenv(key) != "" {
			// Config files have lower precedence.
			continue
		}

		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return nil
}
