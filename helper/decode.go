package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DecodeJSONStrict decodes the request body into dst and rejects unknown fields
// It also rejects a body with trailing content after the JSON value.
func DecodeJSONStrict(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		const prefix = `json: unknown field `
		if strings.HasPrefix(err.Error(), prefix) {
			field := strings.Trim(strings.TrimPrefix(err.Error(), prefix), `"`)
			return fmt.Errorf("unknown field %q", field)
		}
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}

	return nil
}
