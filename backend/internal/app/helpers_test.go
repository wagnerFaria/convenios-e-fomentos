package app_test

import (
	"encoding/json"
	"net/http/httptest"
)

type recorder struct{ *httptest.ResponseRecorder }

func newHTTPRecorder() *httptest.ResponseRecorder { return httptest.NewRecorder() }

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
