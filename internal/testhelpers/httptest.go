package testhelpers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func NewHTTPTestServer(t *testing.T, h http.Handler) *httptest.Server {
	ts := httptest.NewServer(h)
	t.Cleanup(ts.Close)
	return ts
}
