// nolint
package testhelpers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"
	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/internal/httpclient/client/common"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/selfservice/flow/recovery"
)

func NewRecoveryUITestServer(t *testing.T) *httptest.Server {
	router := httprouter.New()
	router.GET("/recovery", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.GET("/settings", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusAccepted)
	})
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	viper.Set(configuration.ViperKeySelfServiceSettingsURL, ts.URL+"/settings")
	viper.Set(configuration.ViperKeySelfServiceRecoveryUI, ts.URL+"/recovery")

	return ts
}

func GetRecoveryRequest(t *testing.T, client *http.Client, ts *httptest.Server) *common.GetSelfServiceBrowserRecoveryRequestOK {
	publicClient := NewSDKClient(ts)

	res, err := client.Get(ts.URL + recovery.PublicRecoveryInitPath)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	rs, err := publicClient.Common.GetSelfServiceBrowserRecoveryRequest(
		common.NewGetSelfServiceBrowserRecoveryRequestParams().WithHTTPClient(client).
			WithRequest(res.Request.URL.Query().Get("request")),
	)
	require.NoError(t, err, "%s", res.Request.URL.String())
	assert.Empty(t, rs.Payload.Active)

	return rs
}

func RecoverySubmitForm(
	t *testing.T,
	f *models.RequestMethodConfig,
	hc *http.Client,
	values url.Values,
) (string, *common.GetSelfServiceBrowserRecoveryRequestOK) {
	require.NotEmpty(t, f.Action)

	res, err := hc.PostForm(pointerx.StringR(f.Action), values)
	require.NoError(t, err)
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	assert.EqualValues(t, http.StatusNoContent, res.StatusCode, "%s", b)

	assert.Equal(t, viper.GetString(configuration.ViperKeySelfServiceRecoveryUI), res.Request.URL.Scheme+"://"+res.Request.URL.Host+res.Request.URL.Path, "should end up at the settings URL, used: %s", pointerx.StringR(f.Action))

	rs, err := NewSDKClientFromURL(viper.GetString(configuration.ViperKeyPublicBaseURL)).Common.GetSelfServiceBrowserRecoveryRequest(
		common.NewGetSelfServiceBrowserRecoveryRequestParams().WithHTTPClient(hc).
			WithRequest(res.Request.URL.Query().Get("request")),
	)
	require.NoError(t, err)
	body, err := json.Marshal(rs.Payload)
	require.NoError(t, err)
	return string(body), rs
}
