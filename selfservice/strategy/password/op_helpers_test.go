// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	hydraclientgo "github.com/ory/hydra-client-go/v2"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/resilience"
	"github.com/ory/x/urlx"

	"github.com/phayes/freeport"
)

type clientAppConfig struct {
	client      *oauth2.Config
	expectToken bool
	state       *clientAppState
}

type clientAppState struct {
	visits int64
	tokens int64
}

type callTrace string

const (
	RegistrationUI                       callTrace = "registration-ui"
	RegistrationWithOAuth2LoginChallenge callTrace = "registration-with-oauth2-login-challenge"
	RegistrationWithFlowID               callTrace = "registration-with-flow-id"
	LoginUI                              callTrace = "login-ui"
	LoginWithOAuth2LoginChallenge        callTrace = "login-with-oauth2-login-challenge"
	LoginWithFlowID                      callTrace = "login-with-flow-id"
	Consent                              callTrace = "consent"
	ConsentWithChallenge                 callTrace = "consent-with-challenge"
	ConsentAccept                        callTrace = "consent-accept"
	ConsentSkip                          callTrace = "consent-skip"
	ConsentClientSkip                    callTrace = "consent-client-skip"
	CodeExchange                         callTrace = "code-exchange"
	CodeExchangeWithToken                callTrace = "code-exchange-with-token"
)

type testContextKey string

const (
	TestUIConfig         testContextKey = "test-ui-config"
	TestOAuthClientState testContextKey = "test-oauth-client-state"
)

type testConfig struct {
	identifier       string
	password         string
	browserClient    *http.Client
	kratosPublicTS   *httptest.Server
	clientAppTS      *httptest.Server
	hydraAdminClient hydraclientgo.OAuth2API
	consentRemember  bool
	requestedScope   []string
	callTrace        *[]callTrace
}

func createHydraOAuth2ApiClient(url string) hydraclientgo.OAuth2API {
	configuration := hydraclientgo.NewConfiguration()
	configuration.Host = urlx.ParseOrPanic(url).Host
	configuration.Servers = hydraclientgo.ServerConfigurations{{URL: url}}

	return hydraclientgo.NewAPIClient(configuration).OAuth2API
}

func createOAuth2Client(t *testing.T, ctx context.Context, hydraAdmin hydraclientgo.OAuth2API, redirectURIs []string, scope string, skipConsent bool) string {
	t.Helper()

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
			SkipConsent:             &skipConsent,
		},
	).Execute()
	require.NoError(t, err)
	require.Equal(t, r.StatusCode, http.StatusCreated)
	return *c.ClientId
}

func makeAuthCodeURL(t *testing.T, c *oauth2.Config, requestedClaims string, isForced bool) string {
	t.Helper()

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
		// Keep tag in sync with the version in ci.yaml
		Tag: "v2.2.0@sha256:6c0f9195fe04ae16b095417b323881f8c9008837361160502e11587663b37c09",
		Env: []string{
			"DSN=memory",
			fmt.Sprintf("URLS_SELF_ISSUER=http://127.0.0.1:%d/", publicPort),
			"URLS_LOGIN=" + loginUI,
			"URLS_CONSENT=" + consentUI,
			"LOG_LEAK_SENSITIVE_VALUES=true",
			"SECRETS_SYSTEM=someverylongsecretthatis32byteslong",
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
		Stdout:       false,
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

func doOAuthFlow(t *testing.T, ctx context.Context, oauthClient *oauth2.Config, browserClient *http.Client) {
	t.Helper()

	authCodeURL := makeAuthCodeURL(t, oauthClient, "", false)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authCodeURL, nil)
	require.NoError(t, err)
	res, err := browserClient.Do(req)
	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.NoError(t, res.Body.Close())
	require.Equal(t, "", string(body))
	require.Equal(t, http.StatusOK, res.StatusCode)
}
