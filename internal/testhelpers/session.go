package testhelpers

import (
	"net/http"
	"testing"

	"github.com/ory/kratos/session"
)

func NewSessionClient(t *testing.T, u string) *http.Client {
	c := session.MockCookieClient(t)
	session.MockHydrateCookieClient(t, c, u)
	return c
}
