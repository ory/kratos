package testhelpers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/internal/httpclient/client/common"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
)

func NewLoginUIFlowEchoServer(t *testing.T, reg driver.Registry) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.LoginFlowPersister().GetLoginFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	viper.Set(configuration.ViperKeySelfServiceLoginUI, ts.URL+"/login-ts")
	t.Cleanup(ts.Close)
	return ts
}

func NewRegistrationUIFlowEchoServer(t *testing.T, reg driver.Registry) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	viper.Set(configuration.ViperKeySelfServiceRegistrationUI, ts.URL+"/registration-ts")
	t.Cleanup(ts.Close)
	return ts
}

func NewSettingsUIFlowEchoServer(t *testing.T, reg driver.Registry) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.SettingsFlowPersister().GetSettingsFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("flow")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
	viper.Set(configuration.ViperKeySelfServiceSettingsURL, ts.URL+"/settings-ts")
	t.Cleanup(ts.Close)
	return ts
}

func InitializeLoginFlowViaBrowser(t *testing.T, client *http.Client, ts *httptest.Server) *common.GetSelfServiceLoginFlowOK {
	publicClient := NewSDKClient(ts)

	res, err := client.Get(ts.URL + login.RouteInitBrowserFlow)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, err := publicClient.Common.GetSelfServiceLoginFlow(
		common.NewGetSelfServiceLoginFlowParams().WithHTTPClient(client).
			WithID(res.Request.URL.Query().Get("flow")),
	)
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func InitializeLoginFlowViaAPI(t *testing.T, client *http.Client, ts *httptest.Server) *common.InitializeSelfServiceLoginViaAPIFlowOK {
	publicClient := NewSDKClient(ts)

	rs, err := publicClient.Common.InitializeSelfServiceLoginViaAPIFlow(common.
		NewInitializeSelfServiceLoginViaAPIFlowParams().WithHTTPClient(client))
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}
