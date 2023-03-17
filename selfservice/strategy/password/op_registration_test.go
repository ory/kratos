// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password_test

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2"

	hydraclientgo "github.com/ory/hydra-client-go/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/x"
)

type clientAppConfig struct {
	client      *oauth2.Config
	expectToken bool
	state       clientAppState
}

type clientAppState struct {
	visits int64
	tokens int64
}

type kratosUIConfig struct {
	expectLoginScreen bool
	identifier        string
	password          string
	browserClient     *http.Client
	kratosPublicTS    *httptest.Server
	clientAppTS       *httptest.Server
	hydraAdminClient  hydraclientgo.OAuth2Api
}

func newClientAppTS(t *testing.T, c *clientAppConfig) *httptest.Server {
	return testhelpers.NewHTTPTestServer(t, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		c.state.visits += 1
		t.Logf("[clientAppTS] handling a callback at client app %s", r.URL.String())
		if r.URL.Query().Has("code") {
			token, err := c.client.Exchange(r.Context(), r.URL.Query().Get("code"))
			require.NoError(t, err)
			require.NotNil(t, token)
			require.NotEqual(t, "", token.AccessToken)
			require.True(t, c.expectToken)
			c.state.tokens += 1
			t.Log("[clientAppTS] successfully exchanged code for token")
		} else {
			t.Error("[clientAppTS] code query parameter is missing")
			require.False(t, c.expectToken)
		}
	}))
}

func kratosUIHandleConsent(t *testing.T, req *http.Request, client *http.Client, haa hydraclientgo.OAuth2Api, clientAppURL string) {
	q := req.URL.Query()
	cr, resp, err := haa.GetOAuth2ConsentRequest(req.Context()).ConsentChallenge(q.Get("consent_challenge")).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.ElementsMatch(t, cr.RequestedScope, []string{"profile", "email"})

	remember := true
	completedAcceptRequest, resp, err := haa.AcceptOAuth2ConsentRequest(context.Background()).AcceptOAuth2ConsentRequest(hydraclientgo.AcceptOAuth2ConsentRequest{
		Remember: &remember,
	}).ConsentChallenge(q.Get("consent_challenge")).Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, completedAcceptRequest)

	t.Logf("[uiTS] navigating to %s", completedAcceptRequest.RedirectTo)
	resp, err = client.Get(completedAcceptRequest.RedirectTo)
	require.NoError(t, err)
	require.Equal(t, clientAppURL, fmt.Sprintf("%s://%s", resp.Request.URL.Scheme, resp.Request.URL.Host))
	require.True(t, resp.Request.URL.Query().Has("code"))
}

func newKratosUITS(t *testing.T, c *kratosUIConfig) *httptest.Server {
	return testhelpers.NewHTTPTestServer(t, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Logf("[uiTS] handling %s", r.URL)
		q := r.URL.Query()

		if len(q) == 1 && !q.Has("flow") && q.Has("login_challenge") {
			t.Log("[uiTS] initializing a new OpenID Provider flow")
			hlc := r.URL.Query().Get("login_challenge")
			f := testhelpers.InitializeRegistrationFlowViaBrowser(t, c.browserClient, c.kratosPublicTS, false, false, !c.expectLoginScreen, testhelpers.InitFlowWithOAuth2LoginChallenge(hlc))
			if c.expectLoginScreen {
				require.NotNil(t, f)

				values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
				values.Set("traits.foobar", c.identifier)
				values.Set("traits.username", c.identifier)
				values.Set("password", c.password)

				_, res := testhelpers.RegistrationMakeRequest(t, false, false, f, c.browserClient, values.Encode())

				assert.EqualValues(t, http.StatusOK, res.StatusCode)
			} else {
				require.Nil(t, f, "registration flow should have been skipped and invalidated, but we successfully retrieved it")
			}
			return
		}

		if q.Has("consent_challenge") {
			kratosUIHandleConsent(t, r, c.browserClient, c.hydraAdminClient, c.clientAppTS.URL)
			return
		}

		if q.Has("flow") {
			t.Log("[uiTS] no operaton; the flow should be completed by the handler that initialized it")
			return
		}

		t.Errorf("[uiTS] unexpected query %#v", q)
	}))
}

func TestOAuth2ProviderRegistration(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	kratosPublicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, x.NewRouterPublic(), x.NewRouterAdmin())
	errTS := testhelpers.NewErrorTestServer(t, reg)
	redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

	var hydraAdminClient hydraclientgo.OAuth2Api

	cac := &clientAppConfig{}
	clientAppTS := newClientAppTS(t, cac)

	kuc := &kratosUIConfig{}
	kratosUITS := newKratosUITS(t, kuc)

	hydraAdmin, hydraPublic := newHydra(t, kratosUITS.URL, kratosUITS.URL)

	hydraAdminClient = createHydraOAuth2ApiClient(hydraAdmin)
	clientID := createOAuth2Client(t, ctx, hydraAdminClient, []string{clientAppTS.URL}, "profile email")

	defaultClient := &oauth2.Config{
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

	conf.MustSet(ctx, config.ViperKeyOAuth2ProviderURL, hydraAdmin+"/")
	conf.MustSet(ctx, config.ViperKeySelfServiceErrorUI, errTS.URL+"/error-ts")
	conf.MustSet(ctx, config.ViperKeySelfServiceLoginUI, kratosUITS.URL+"/login-ts")
	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationUI, kratosUITS.URL+"/login-ts")
	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, redirTS.URL+"/return-ts")
	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+"."+config.DefaultBrowserReturnURL, redirTS.URL+"/registration-return-ts")
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{"enabled": true})
	conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})
	conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), []config.SelfServiceHook{{Name: "session"}})
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")

	sharedBrowserClient := testhelpers.NewClientWithCookieJar(t, nil, true)

	type state struct {
		cas clientAppState
	}
	for _, tc := range []struct {
		name      string
		configure func(c *config.Config)
		cac       clientAppConfig
		kuc       kratosUIConfig
		expected  state
	}{
		{
			name: "should prompt the user for login and consent",
			configure: func(c *config.Config) {
				c.MustSet(ctx, config.ViperKeySessionPersistentCookie, false)
			},
			cac: clientAppConfig{
				client:      defaultClient,
				expectToken: true,
			},
			kuc: kratosUIConfig{
				expectLoginScreen: true,
				identifier:        x.NewUUID().String(),
				password:          x.NewUUID().String(),
				browserClient:     sharedBrowserClient,
				kratosPublicTS:    kratosPublicTS,
				clientAppTS:       clientAppTS,
				hydraAdminClient:  hydraAdminClient,
			},
			expected: state{
				cas: clientAppState{
					visits: 1,
					tokens: 1,
				},
			},
		},
		{
			name: "should prompt the user for login and consent again",
			configure: func(c *config.Config) {
				c.MustSet(ctx, config.ViperKeySessionPersistentCookie, true)
			},
			cac: clientAppConfig{
				client:      defaultClient,
				expectToken: true,
			},
			kuc: kratosUIConfig{
				expectLoginScreen: true,
				identifier:        x.NewUUID().String(),
				password:          x.NewUUID().String(),
				browserClient:     sharedBrowserClient,
				kratosPublicTS:    kratosPublicTS,
				clientAppTS:       clientAppTS,
				hydraAdminClient:  hydraAdminClient,
			},
			expected: state{
				cas: clientAppState{
					visits: 1,
					tokens: 1,
				},
			},
		},
		{
			name: "should fail because the persistent Hydra session doesn't match the new Kratos session subject",
			configure: func(c *config.Config) {
			},
			cac: clientAppConfig{
				client:      defaultClient,
				expectToken: false,
			},
			kuc: kratosUIConfig{
				expectLoginScreen: true,
				identifier:        x.NewUUID().String(),
				password:          x.NewUUID().String(),
				browserClient:     sharedBrowserClient,
				kratosPublicTS:    kratosPublicTS,
				clientAppTS:       clientAppTS,
				hydraAdminClient:  hydraAdminClient,
			},
			expected: state{
				cas: clientAppState{
					visits: 0,
					tokens: 0,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			*cac = tc.cac
			*kuc = tc.kuc
			tc.configure(conf)

			authCodeURL := makeAuthCodeURL(t, cac.client, "", false)
			res, err := tc.kuc.browserClient.Get(authCodeURL)

			require.NoError(t, err)
			body, err := io.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)
			require.Equal(t, "", string(body))
			require.Equal(t, http.StatusOK, res.StatusCode)
			require.EqualValues(t, tc.expected.cas, cac.state)
		})
	}
}
