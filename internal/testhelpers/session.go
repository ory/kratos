package testhelpers

import (
	"net/http"
	"testing"
)

func NewSessionClient(t *testing.T, u string) *http.Client {
	c := MockCookieClient(t)
	MockHydrateCookieClient(t, c, u)
	return c
}
