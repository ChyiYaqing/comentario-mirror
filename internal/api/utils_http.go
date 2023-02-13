package api

import (
	"encoding/json"
	"gitlab.com/comentario/comentario/internal/util"
	"io"
	"net/http"
	"reflect"
)

type response map[string]interface{}

// TODO: Add tests in utils_http_test.go

func BodyUnmarshal(r *http.Request, x interface{}) error {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("cannot read POST body: %v\n", err)
		return util.ErrorInternal
	}

	if err = json.Unmarshal(b, x); err != nil {
		return util.ErrorInvalidJSONBody
	}

	xv := reflect.Indirect(reflect.ValueOf(x))
	for i := 0; i < xv.NumField(); i++ {
		if xv.Field(i).IsNil() {
			return util.ErrorMissingField
		}
	}

	return nil
}

func BodyMarshal(w http.ResponseWriter, x map[string]interface{}) error {
	resp, err := json.Marshal(x)
	if err != nil {
		_, _ = w.Write([]byte(`{"success":false,"message":"Some internal error occurred"}`))
		logger.Errorf("cannot marshal response: %v\n")
		return util.ErrorInternal
	}

	_, err = w.Write(resp)
	return err
}

func BodyMarshalChecked(w http.ResponseWriter, x map[string]interface{}) {
	if err := BodyMarshal(w, x); err != nil {
		logger.Warningf("failed to write success response ($#v): %v", x, err)
	}
}

func getIP(r *http.Request) string {
	ip := r.RemoteAddr
	if r.Header.Get("X-Forwarded-For") != "" {
		ip = r.Header.Get("X-Forwarded-For")
	}

	return ip
}

func getUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}
