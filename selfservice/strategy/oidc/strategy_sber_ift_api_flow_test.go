package oidc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/x/configx"
	"github.com/ory/x/httprouterx"
	"github.com/ory/x/urlx"
)

func TestSberIFTMockServer_APILoginAndRegistrationFlow(t *testing.T) {
	ctx := context.Background()
	mock := newSberIFTMockServer(t)

	conf, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(map[string]any{
			config.ViperKeyIdentitySchemas: config.Schemas{
				{ID: "default", URL: "file://./stub/registration.schema.json"},
			},
			config.ViperKeyDefaultIdentitySchemaID: "default",
			config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeOIDC.String()): []config.SelfServiceHook{{Name: "session"}},
		}),
	)

	returnTS := newReturnTS(t, reg)
	newUI(t, reg)
	testhelpers.NewErrorTestServer(t, reg)

	routerP, routerA := httprouterx.NewTestRouterPublic(t), httprouterx.NewTestRouterAdminWithPrefix(t)
	kratosTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, routerP, routerA)

	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, returnTS.URL)
	conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnTS.URL})

	viperSetProviderConfig(t, conf, oidc.Configuration{
		Provider:         "sber-ift",
		ID:               "sber-ift",
		ClientID:         "sber-ift-client",
		ClientSecret:     "sber-ift-secret",
		Scope:            []string{"openid", "email"},
		Mapper:           "file://./stub/oidc.hydra.jsonnet",
		AuthURL:          mock.URL() + "/authorize",
		TokenURL:         mock.URL() + "/token",
		UserInfoURL:      mock.URL() + "/userinfo",
		AuthCompletedURL: mock.URL() + "/auth/completed",
	})

	runOIDCAPICallbackFlow(t, kratosTS.URL, registration.RouteInitAPIFlow, registration.RouteSubmitFlow)
	runOIDCAPICallbackFlow(t, kratosTS.URL, login.RouteInitAPIFlow, login.RouteSubmitFlow)

	assert.GreaterOrEqual(t, mock.AuthCompletedCalls(), 2)
}

func runOIDCAPICallbackFlow(t *testing.T, publicURL, initPath, submitPath string) {
	t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := testhelpers.NewClientWithCookieJar(t, jar, func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	})

	req, err := http.NewRequest(http.MethodGet, publicURL+initPath, nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	var initResp map[string]any
	require.NoError(t, json.NewDecoder(res.Body).Decode(&initResp))
	flowID, _ := initResp["id"].(string)
	require.NotEmpty(t, flowID)

	payload := bytes.NewBufferString(`{"method":"oidc","provider":"sber-ift"}`)
	req, err = http.NewRequest(http.MethodPost, publicURL+submitPath+"?flow="+url.QueryEscape(flowID), payload)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	res, err = client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

	var locationErr flow.BrowserLocationChangeRequiredError
	require.NoError(t, json.NewDecoder(res.Body).Decode(&locationErr))
	require.NotEmpty(t, locationErr.RedirectBrowserTo)

	res, err = client.Get(locationErr.RedirectBrowserTo)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusSeeOther, res.StatusCode)

	callbackLocation := res.Header.Get("Location")
	require.NotEmpty(t, callbackLocation)
	require.Contains(t, callbackLocation, "/self-service/methods/oidc/callback/sber-ift")

	callbackURL := callbackLocation
	if strings.HasPrefix(callbackLocation, "/") {
		callbackURL = publicURL + callbackLocation
	}

	callbackReq, err := http.NewRequest(http.MethodGet, callbackURL, nil)
	require.NoError(t, err)
	callbackReq.Header.Set("Accept", "application/json")
	res, err = client.Do(callbackReq)
	require.NoError(t, err)
	defer res.Body.Close()
	callbackBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Contains(t, []int{http.StatusSeeOther, http.StatusOK}, res.StatusCode, "callback body: %s", string(callbackBody))

	if res.StatusCode == http.StatusOK {
		assert.Contains(t, string(callbackBody), `"session_token"`)
		assert.Contains(t, string(callbackBody), `"continue_with"`)
		return
	}
	require.Equal(t, urlx.ParseOrPanic(publicURL).Hostname(), res.Request.URL.Hostname())

	successLocation := res.Header.Get("Location")
	require.NotEmpty(t, successLocation)
	successURL, err := url.Parse(successLocation)
	require.NoError(t, err)
	assert.Equal(t, "/api/auth/oidc/success", successURL.Path)

	cookies := res.Header.Values("Set-Cookie")
	joinedCookies := strings.Join(cookies, "\n")
	assert.Contains(t, joinedCookies, "auth=")
	assert.Contains(t, joinedCookies, "ory_kratos_session")

}
