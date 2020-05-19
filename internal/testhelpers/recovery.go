package testhelpers

import (
	"net/http"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/internal/httpclient/client/common"
	"github.com/ory/kratos/selfservice/flow/settings"
)

func NewRecoveryTestServer(t *testing.T) *httptest.Server {
	router := httprouter.New()
	router.GET("/recovery", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusNoContent)
	})
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	viper.Set(configuration.ViperKeyURLsRecovery, ts.URL+"/recovery")

	return ts
}

func GetRecoveryRequest(t *testing.T, primaryUser *http.Client, ts *httptest.Server) *common.GetSelfServiceBrowserRecoveryRequestOK {
	publicClient := NewSDKClient(ts)

	res, err := primaryUser.Get(ts.URL + settings.PublicPath)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, err := publicClient.Common.GetSelfServiceBrowserRecoveryRequest(
		common.NewGetSelfServiceBrowserRecoveryRequestParams().WithHTTPClient(primaryUser).
			WithRequest(res.Request.URL.Query().Get("request")),
	)
	require.NoError(t, err)
	assert.Empty(t, rs.Payload.Active)

	return rs
}
