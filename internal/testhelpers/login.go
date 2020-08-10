package testhelpers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/x"
)

func NewLoginUIRequestEchoServer(t *testing.T, reg driver.Registry) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.LoginRequestPersister().GetLoginRequest(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	viper.Set(configuration.ViperKeySelfServiceLoginUI, ts.URL+"/login-ts")
	t.Cleanup(ts.Close)
	return ts
}
