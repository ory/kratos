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

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"

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
	expectLoginScreen        bool
	expectRegistrationScreen bool
	identifier               string
	password                 string
	browserClient            *http.Client
	kratosPublicTS           *httptest.Server
	clientAppTS              *httptest.Server
	hydraAdminClient         hydraclientgo.OAuth2Api
}

func TestOAuth2ProviderRegistration(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)

	kratosPublicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, x.NewRouterPublic(), x.NewRouterAdmin())
	errTS := testhelpers.NewErrorTestServer(t, reg)
	redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

	var hydraAdminClient hydraclientgo.OAuth2Api

	router := x.NewRouterPublic()

	const (
		TestUIConfig         = "test-ui-config"
		TestOAuthClientState = "test-oauth-client-state"
	)

	router.GET("/login-ts", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		t.Log("[uiTS] navigated to the login ui")

		q := r.URL.Query()
		hlc := r.URL.Query().Get("login_challenge")

		c := r.Context().Value(TestUIConfig).(*kratosUIConfig)

		require.Truef(t, c.expectLoginScreen, "got the login screen but did not expect it. expectLoginScreen: %t", c.expectLoginScreen)

		if hlc != "" {
			f := testhelpers.InitializeLoginFlowViaBrowser(t, c.browserClient, c.kratosPublicTS, false, false, false, false, testhelpers.InitFlowWithOAuth2LoginChallenge(q.Get("login_challenge")))
			require.NotNil(t, f)

			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)

			if *f.Refresh {
				t.Log("[uiTS] login flow has a refresh flag. Let's supply the user password")
			} else {
				values.Set("traits.foobar", c.identifier)
				values.Set("traits.username", c.identifier)
			}
			values.Set("password", c.password)

			body, res := testhelpers.LoginMakeRequest(t, false, false, f, c.browserClient, values.Encode())
			assert.EqualValues(t, http.StatusOK, res.StatusCode)

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(body))
			return
		}

		if q.Has("flow") {
			t.Log("[uiTS] login flow is ignored here since it will be handled by the code above, we just need to return")
			w.WriteHeader(http.StatusOK)
			return
		}
	})

	router.GET("/registration-ts", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		t.Log("[uiTS] navigated to the registration ui")
		c := r.Context().Value(TestUIConfig).(*kratosUIConfig)

		q := r.URL.Query()
		hlc := q.Get("login_challenge")

		if hlc != "" {
			t.Log("[uiTS] initializing a new OpenID Provider flow through the registration endpoint")
			f := testhelpers.InitializeRegistrationFlowViaBrowser(t, c.browserClient, c.kratosPublicTS, false, false, false, testhelpers.InitFlowWithOAuth2LoginChallenge(hlc))
			require.NotNilf(t, f, "expected a flow to be initialized but got none")

			// continue the registration flow here
			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			values.Set("traits.foobar", c.identifier)
			values.Set("traits.username", c.identifier)
			values.Set("password", c.password)

			body, res := testhelpers.RegistrationMakeRequest(t, false, false, f, c.browserClient, values.Encode())
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(body))
			return
		}

		if q.Has("flow") {
			require.Truef(t, c.expectRegistrationScreen, "got the registration screen but did not expect it. expectRegistrationScreen: %t", c.expectRegistrationScreen)
			t.Log("[uiTS] registration flow is ignored here since it will be handled by the code above, we just need to return")
			w.WriteHeader(http.StatusOK)
			return
		}
	})

	router.GET("/consent", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		q := r.URL.Query()

		c := r.Context().Value(TestUIConfig).(*kratosUIConfig)
		consentChallenge := q.Get("consent_challenge")
		assert.NotEmpty(t, consentChallenge)

		if consentChallenge == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cr, resp, err := hydraAdminClient.GetOAuth2ConsentRequest(ctx).ConsentChallenge(q.Get("consent_challenge")).Execute()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.ElementsMatch(t, cr.RequestedScope, []string{"profile", "email"})

		remember := true
		completedAcceptRequest, resp, err := hydraAdminClient.AcceptOAuth2ConsentRequest(context.Background()).AcceptOAuth2ConsentRequest(hydraclientgo.AcceptOAuth2ConsentRequest{
			Remember: &remember,
		}).ConsentChallenge(q.Get("consent_challenge")).Execute()

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NotNil(t, completedAcceptRequest)

		t.Logf("[uiTS] navigating to %s", completedAcceptRequest.RedirectTo)
		resp, err = c.browserClient.Get(completedAcceptRequest.RedirectTo)
		require.NoError(t, err)
		require.Equal(t, c.clientAppTS.URL, fmt.Sprintf("%s://%s", resp.Request.URL.Scheme, resp.Request.URL.Host))
		require.True(t, resp.Request.URL.Query().Has("code"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	})

	router.GET("/", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		c := r.Context().Value(TestUIConfig).(*kratosUIConfig)
		q := r.URL.Query()
		hlc := q.Get("login_challenge")

		if hlc != "" {
			t.Log("[uiTS] initializing a new OpenID Provider flow through the registration endpoint")
			f := testhelpers.InitializeRegistrationFlowViaBrowser(t, c.browserClient, c.kratosPublicTS, false, false, false, testhelpers.InitFlowWithOAuth2LoginChallenge(hlc))
			require.NotNilf(t, f, "expected a flow to be initialized but got none")

			// continue the registration flow here
			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			values.Set("traits.foobar", c.identifier)
			values.Set("traits.username", c.identifier)
			values.Set("password", c.password)

			body, res := testhelpers.RegistrationMakeRequest(t, false, false, f, c.browserClient, values.Encode())
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(body))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	})

	kratosUIMiddleware := negroni.New()
	kratosUIMiddleware.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		// add the context from the global context to each request
		next(rw, r.WithContext(ctx))
	})
	kratosUIMiddleware.UseHandler(router)

	kratosUITS := testhelpers.NewHTTPTestServer(t, kratosUIMiddleware)

	clientAppTSMiddleware := negroni.New()
	clientAppTSMiddleware.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		// add the context from the global context to each request
		next(rw, r.WithContext(ctx))
	})
	clientAppTSMiddleware.UseHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := ctx.Value(TestOAuthClientState).(*clientAppConfig)
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
		w.WriteHeader(http.StatusOK)
	})
	// A new OAuth client which will also function as the callback for the code exchange
	clientAppTS := testhelpers.NewHTTPTestServer(t, clientAppTSMiddleware)

	// we want to test if the registration ui is used if the flow contains an oauth2 login challenge
	// so we will have Hydra redirect to the base path of the test kratos ui server which
	// will then initiate the registration flow
	hydraAdmin, hydraPublic := newHydra(t, kratosUITS.URL+"/registration-ts", kratosUITS.URL+"/consent")

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
	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationUI, kratosUITS.URL+"/registration-ts")
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
	identifier := x.NewUUID().String()
	password := x.NewUUID().String()

	doOAuthFlow := func(t *testing.T, ctx context.Context, expected state) {
		t.Helper()

		authCodeURL := makeAuthCodeURL(t, defaultClient, "", false)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, authCodeURL, nil)
		require.NoError(t, err)

		res, err := sharedBrowserClient.Do(req)
		require.NoError(t, err)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		require.NoError(t, res.Body.Close())
		require.Equal(t, "", string(body))
		require.Equal(t, http.StatusOK, res.StatusCode)

		cac := ctx.Value(TestOAuthClientState).(*clientAppConfig)
		require.EqualValues(t, expected.cas, cac.state)
	}

	t.Run("case=should accept oauth login request on registration", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySessionPersistentCookie, false)

		ctx = context.WithValue(ctx, TestUIConfig, &kratosUIConfig{
			expectLoginScreen:        false,
			expectRegistrationScreen: true,
			identifier:               identifier,
			password:                 password,
			browserClient:            sharedBrowserClient,
			kratosPublicTS:           kratosPublicTS,
			clientAppTS:              clientAppTS,
			hydraAdminClient:         hydraAdminClient,
		})

		ctx = context.WithValue(ctx, TestOAuthClientState, &clientAppConfig{
			client:      defaultClient,
			expectToken: true,
		})

		doOAuthFlow(t, ctx, state{
			cas: clientAppState{
				visits: 1,
				tokens: 1,
			},
		})
	})

	t.Run("case=should prompt the user for oauth consent even with session", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySessionPersistentCookie, true)

		ctx = context.WithValue(ctx, TestUIConfig, &kratosUIConfig{
			expectLoginScreen:        true,
			expectRegistrationScreen: false,
			identifier:               identifier,
			password:                 password,
			browserClient:            sharedBrowserClient,
			kratosPublicTS:           kratosPublicTS,
			clientAppTS:              clientAppTS,
			hydraAdminClient:         hydraAdminClient,
		})

		ctx = context.WithValue(ctx, TestOAuthClientState, &clientAppConfig{
			client:      defaultClient,
			expectToken: true,
		})

		doOAuthFlow(t, ctx, state{
			cas: clientAppState{
				visits: 1,
				tokens: 1,
			},
		})
	})

	t.Run("case=should fail because the persistent Hydra session doesn't match the new Kratos session subject", func(t *testing.T) {
		ctx = context.WithValue(ctx, TestUIConfig, &kratosUIConfig{
			expectLoginScreen:        true,
			expectRegistrationScreen: false,
			identifier:               x.NewUUID().String(),
			password:                 x.NewUUID().String(),
			kratosPublicTS:           kratosPublicTS,
			browserClient:            sharedBrowserClient,
			clientAppTS:              clientAppTS,
			hydraAdminClient:         hydraAdminClient,
		})

		ctx = context.WithValue(ctx, TestOAuthClientState, &clientAppConfig{
			client:      defaultClient,
			expectToken: false,
		})

		doOAuthFlow(t, ctx, state{
			cas: clientAppState{
				visits: 0,
				tokens: 0,
			},
		})
	})
}
