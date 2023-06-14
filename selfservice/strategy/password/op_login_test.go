// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password_test

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/gofrs/uuid"

	"github.com/ory/x/logrusx"
	"github.com/ory/x/resilience"
	"github.com/ory/x/urlx"

	hydraclientgo "github.com/ory/hydra-client-go/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func createHydraOAuth2ApiClient(url string) hydraclientgo.OAuth2Api {
	configuration := hydraclientgo.NewConfiguration()
	configuration.Host = urlx.ParseOrPanic(url).Host
	configuration.Servers = hydraclientgo.ServerConfigurations{{URL: url}}

	return hydraclientgo.NewAPIClient(configuration).OAuth2Api
}

func createOAuth2Client(t *testing.T, ctx context.Context, hydraAdmin hydraclientgo.OAuth2Api, redirectURIs []string, scope string) string {
	clientName := "kratos-hydra-integration-test-client-1"
	tokenEndpointAuthMethod := "client_secret_post"
	clientSecret := "client-secret"

	c, r, err := hydraAdmin.CreateOAuth2Client(ctx).OAuth2Client(
		hydraclientgo.OAuth2Client{
			ClientName:              &clientName,
			RedirectUris:            redirectURIs,
			Scope:                   &scope,
			TokenEndpointAuthMethod: &tokenEndpointAuthMethod,
			ClientSecret:            &clientSecret,
		},
	).Execute()
	require.NoError(t, err)
	require.Equal(t, r.StatusCode, http.StatusCreated)
	return *c.ClientId
}

func makeAuthCodeURL(t *testing.T, c *oauth2.Config, requestedClaims string, isForced bool) string {
	var options []oauth2.AuthCodeOption

	if isForced {
		options = append(options, oauth2.SetAuthURLParam("prompt", "login"))
	}
	if requestedClaims != "" {
		options = append(options, oauth2.SetAuthURLParam("claims", requestedClaims))
	}

	state := fmt.Sprintf("%x", uuid.Must(uuid.NewV4()))
	return c.AuthCodeURL(state, options...)
}

func newHydra(t *testing.T, loginUI string, consentUI string) (hydraAdmin string, hydraPublic string) {
	publicPort, err := freeport.GetFreePort()
	require.NoError(t, err)
	adminPort, err := freeport.GetFreePort()
	require.NoError(t, err)

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	hydraResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "oryd/hydra",
		Tag:        "v2.0.0",
		Env: []string{
			"DSN=memory",
			fmt.Sprintf("URLS_SELF_ISSUER=http://127.0.0.1:%d/", publicPort),
			"URLS_LOGIN=" + loginUI,
			"URLS_CONSENT=" + consentUI,
			"LOG_LEAK_SENSITIVE_VALUES=true",
		},
		Cmd:          []string{"serve", "all", "--dev"},
		ExposedPorts: []string{"4444/tcp", "4445/tcp"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"4444/tcp": {{HostPort: fmt.Sprintf("%d/tcp", publicPort)}},
			"4445/tcp": {{HostPort: fmt.Sprintf("%d/tcp", adminPort)}},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, hydraResource.Close())
	})

	require.NoError(t, hydraResource.Expire(uint(60*5)))

	require.NotEmpty(t, hydraResource.GetPort("4444/tcp"), "%+v", hydraResource.Container.NetworkSettings.Ports)
	require.NotEmpty(t, hydraResource.GetPort("4445/tcp"), "%+v", hydraResource.Container)

	hydraPublic = "http://127.0.0.1:" + hydraResource.GetPort("4444/tcp")
	hydraAdmin = "http://127.0.0.1:" + hydraResource.GetPort("4445/tcp")

	go pool.Client.Logs(docker.LogsOptions{
		ErrorStream:  TestLogWriter{T: t, streamName: "hydra-stderr"},
		OutputStream: TestLogWriter{T: t, streamName: "hydra-stdout"},
		Stdout:       true,
		Stderr:       true,
		Follow:       true,
		Container:    hydraResource.Container.ID,
	})
	hl := logrusx.New("hydra-ready-check", "hydra-ready-check")
	err = resilience.Retry(hl, time.Second*1, time.Second*5, func() error {
		pr := hydraPublic + "/health/ready"
		res, err := http.DefaultClient.Get(pr)
		if err != nil || res.StatusCode != 200 {
			return errors.Errorf("Hydra public is not ready at " + pr)
		}

		ar := hydraAdmin + "/health/ready"
		res, err = http.DefaultClient.Get(ar)
		if err != nil && res.StatusCode != 200 {
			return errors.Errorf("Hydra admin is not ready at " + ar)
		} else {
			return nil
		}
	})
	require.NoError(t, err)

	t.Logf("Ory Hydra running at: %s %s", hydraPublic, hydraAdmin)

	return hydraAdmin, hydraPublic
}

type TestLogWriter struct {
	streamName string
	*testing.T
}

func (t TestLogWriter) Write(p []byte) (int, error) {
	t.Logf("[%d bytes @ %s]:\n\n%s\n", len(p), t.streamName, string(p))
	return len(p), nil
}

func TestOAuth2Provider(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(
		ctx,
		config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword),
		map[string]interface{}{"enabled": true},
	)

	router := x.NewRouterPublic()
	kratosPublicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	browserClient := testhelpers.NewClientWithCookieJar(t, nil, true)

	errTS := testhelpers.NewErrorTestServer(t, reg)

	redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

	oAuthSuccess := false
	var hydraAdminClient hydraclientgo.OAuth2Api
	var clientAppOAuth2Config *oauth2.Config

	clientAppTS := testhelpers.NewHTTPTestServer(t, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Logf("[clientAppTS] handling a callback at client app %s", r.URL.String())
		if r.URL.Query().Has("code") {
			token, err := clientAppOAuth2Config.Exchange(r.Context(), r.URL.Query().Get("code"))
			require.NoError(t, err)
			require.NotNil(t, token)
			require.NotEqual(t, "", token.AccessToken)
			oAuthSuccess = true
			t.Log("[clientAppTS] successfully exchanged code for token")
		} else {
			t.Error("[clientAppTS] code query parameter is missing")
		}
	}))

	identifier, pwd := x.NewUUID().String(), "password"

	testRequireLogin := true

	uiTS := testhelpers.NewHTTPTestServer(t, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Logf("[uiTS] handling %s", r.URL)
		q := r.URL.Query()

		if len(q) == 1 && !q.Has("flow") && q.Has("login_challenge") {
			t.Log("[uiTS] initializing a new OpenID Provider flow")
			hlc := r.URL.Query().Get("login_challenge")
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, kratosPublicTS, false, false, false, !testRequireLogin, testhelpers.InitFlowWithOAuth2LoginChallenge(hlc))
			if testRequireLogin {
				require.NotNil(t, f)

				values := url.Values{"method": {"password"}, "identifier": {identifier}, "password": {pwd}, "csrf_token": {x.FakeCSRFToken}}.Encode()
				_, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, values)

				assert.EqualValues(t, http.StatusOK, res.StatusCode)
			} else {
				require.Nil(t, f, "login flow should have been skipped and invalidated, but we successfully retrieved it")
			}
			return
		}

		if q.Has("consent_challenge") {
			kratosUIHandleConsent(t, r, browserClient, hydraAdminClient, clientAppTS.URL)
			return
		}

		if q.Has("flow") {
			t.Log("[uiTS] no operaton; the flow should be completed by the handler that initialized it")
			return
		}

		t.Errorf("[uiTS] unexpected query %#v", q)
	}))

	conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})
	conf.MustSet(ctx, config.ViperKeySelfServiceErrorUI, errTS.URL+"/error-ts")
	conf.MustSet(ctx, config.ViperKeySelfServiceLoginUI, uiTS.URL+"/login-ts")
	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, redirTS.URL+"/return-ts")

	testhelpers.SetDefaultIdentitySchemaFromRaw(conf, loginSchema)
	createIdentity(ctx, reg, t, identifier, pwd)

	hydraAdmin, hydraPublic := newHydra(t, uiTS.URL, uiTS.URL)
	conf.MustSet(ctx, config.ViperKeyOAuth2ProviderURL, hydraAdmin)

	hydraAdminClient = createHydraOAuth2ApiClient(hydraAdmin)
	clientID := createOAuth2Client(t, ctx, hydraAdminClient, []string{clientAppTS.URL}, "profile email")

	t.Run("should sign in the user without OAuth2", func(t *testing.T) {
		f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, kratosPublicTS, false, false, false, false)

		values := url.Values{"method": {"password"}, "identifier": {identifier}, "password": {pwd}, "csrf_token": {x.FakeCSRFToken}}.Encode()

		body, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, values)

		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, identifier, gjson.Get(body, "identity.traits.subject").String(), "%s", body)
	})

	clientAppOAuth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: "client-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:   hydraPublic + "/oauth2/auth",
			TokenURL:  hydraPublic + "/oauth2/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes:      []string{"profile", "email"},
		RedirectURL: clientAppTS.URL,
	}

	conf.MustSet(ctx, config.ViperKeySessionPersistentCookie, false)
	t.Run("should prompt the user for login and consent", func(t *testing.T) {
		authCodeURL := makeAuthCodeURL(t, clientAppOAuth2Config, "", false)
		res, err := browserClient.Get(authCodeURL)

		require.NoError(t, err, authCodeURL)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		require.NoError(t, err)
		require.Equal(t, "", string(body))
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.True(t, oAuthSuccess)
		oAuthSuccess = false
	})

	conf.MustSet(ctx, config.ViperKeySessionPersistentCookie, true)
	t.Run("should prompt the user for login and consent again", func(t *testing.T) {
		authCodeURL := makeAuthCodeURL(t, clientAppOAuth2Config, "", false)
		res, err := browserClient.Get(authCodeURL)

		require.NoError(t, err, authCodeURL)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		require.NoError(t, err)
		require.Equal(t, "", string(body))
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.True(t, oAuthSuccess)
		oAuthSuccess = false
	})

	testRequireLogin = false
	t.Run("should prompt the user for consent, but not for login", func(t *testing.T) {
		authCodeURL := makeAuthCodeURL(t, clientAppOAuth2Config, "", false)
		res, err := browserClient.Get(authCodeURL)

		require.NoError(t, err, authCodeURL)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		require.NoError(t, err)
		require.Equal(t, "", string(body))
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.True(t, oAuthSuccess)
		oAuthSuccess = false
	})

	reg.WithHydra(&AcceptWrongSubject{h: reg.Hydra().(*hydra.DefaultHydra)})
	t.Run("should fail when Hydra session subject doesn't match the subject authenticated by Kratos", func(t *testing.T) {
		authCodeURL := makeAuthCodeURL(t, clientAppOAuth2Config, "", false)
		res, err := browserClient.Get(authCodeURL)

		require.NoError(t, err, authCodeURL)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		require.NoError(t, err)
		require.Equal(t, "", string(body))
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.False(t, oAuthSuccess)
		oAuthSuccess = false
	})
}

var _ hydra.Hydra = &AcceptWrongSubject{}

type AcceptWrongSubject struct {
	h *hydra.DefaultHydra
}

func (h *AcceptWrongSubject) AcceptLoginRequest(ctx context.Context, loginChallenge string, subject string, amr session.AuthenticationMethods) (string, error) {
	hackerman := uuid.Must(uuid.NewV4())
	return h.h.AcceptLoginRequest(ctx, loginChallenge, hackerman.String(), amr)
}

func (h *AcceptWrongSubject) GetLoginRequest(ctx context.Context, loginChallenge string) (*hydraclientgo.OAuth2LoginRequest, error) {
	return h.h.GetLoginRequest(ctx, loginChallenge)
}
