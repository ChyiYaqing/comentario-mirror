package svc

import (
	"bytes"
	"encoding/json"
	"gitlab.com/comentario/comentario/internal/config"
	"io"
	"net/http"
	"net/url"
	"time"
)

// TheVersionCheckService is a global VersionCheckService implementation
var TheVersionCheckService VersionCheckService = &versionCheckService{}

// VersionCheckService is a service that periodically checks for version updates
type VersionCheckService interface {
	Init()
}

type versionCheckService struct{}

func (svc *versionCheckService) Init() {
	svc.run()
}

func (svc *versionCheckService) run() {
	go func() {
		printedError := false
		errorCount := 0
		latestSeen := ""

		for {
			time.Sleep(5 * time.Minute)

			data := url.Values{
				"version": {config.AppVersion},
			}

			var body []byte
			var err error
			func() {
				var resp *http.Response
				// TODO version.comentario.io doesn't exist yet
				resp, err = http.Post("https://version.comentario.io/api/check", "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
				if err != nil {
					// Print the error only once; we don't want to spam the logs with this every five minutes
					if !printedError && errorCount > 5 {
						logger.Errorf("error checking version: %v", err)
						printedError = true
					}
				} else {
					defer resp.Body.Close()
					body, err = io.ReadAll(resp.Body)
				}
			}()
			if err != nil {
				errorCount++
				if !printedError && errorCount > 5 {
					logger.Errorf("error reading body: %s", err)
					printedError = true
				}
				continue
			}

			type response struct {
				Success   bool   `json:"success"`
				Message   string `json:"message"`
				Latest    string `json:"latest"`
				NewUpdate bool   `json:"newUpdate"`
			}

			r := response{}
			err = json.Unmarshal(body, &r)
			if err != nil || !r.Success {
				errorCount++
				if !printedError && errorCount > 5 {
					logger.Errorf("error checking version: %s", r.Message)
					printedError = true
				}
				continue
			}

			if r.NewUpdate && r.Latest != latestSeen {
				logger.Infof("New update available! Latest version: %s", r.Latest)
				latestSeen = r.Latest
			}

			errorCount = 0
			printedError = false
		}
	}()
}
