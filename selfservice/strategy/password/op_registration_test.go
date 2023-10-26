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
	"net/url"
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
	"github.com/ory/kratos/selfservice/flow/registration"
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
	CodeExchange                         callTrace = "code-exchange"
	CodeExchangeWithToken                callTrace = "code-exchange-with-token"
)

type kratosUIConfig struct {
	identifier       string
	password         string
	browserClient    *http.Client
	kratosPublicTS   *httptest.Server
	clientAppTS      *httptest.Server
	hydraAdminClient hydraclientgo.OAuth2Api
	consentRemember  bool
	callTrace        []callTrace
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
		t.Log("[loginTS] navigated to the login ui")
		c := r.Context().Value(TestUIConfig).(*kratosUIConfig)
		c.callTrace = append(c.callTrace, LoginUI)

		q := r.URL.Query()
		hlc := r.URL.Query().Get("login_challenge")

		if hlc != "" {
			c.callTrace = append(c.callTrace, LoginWithOAuth2LoginChallenge)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, c.browserClient, c.kratosPublicTS, false, false, false, false, testhelpers.InitFlowWithOAuth2LoginChallenge(q.Get("login_challenge")))

			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			values.Set("password", c.password)

			_, _ = testhelpers.LoginMakeRequest(t, false, false, f, c.browserClient, values.Encode())
			return
		}

		if q.Has("flow") {
			c.callTrace = append(c.callTrace, LoginWithFlowID)
			t.Log("[loginTS] login flow is ignored here since it will be handled by the code above, we just need to return")
			return
		}
	})

	router.GET("/registration-ts", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		t.Log("[registrationTS] navigated to the registration ui")
		c := r.Context().Value(TestUIConfig).(*kratosUIConfig)
		c.callTrace = append(c.callTrace, RegistrationUI)

		q := r.URL.Query()
		hlc := q.Get("login_challenge")

		if hlc != "" {
			c.callTrace = append(c.callTrace, RegistrationWithOAuth2LoginChallenge)
			t.Log("[registrationTS] initializing a new OpenID Provider flow through the registration endpoint")
			registrationUrl, err := url.Parse(c.kratosPublicTS.URL + registration.RouteInitBrowserFlow)
			require.NoError(t, err)

			q := registrationUrl.Query()
			q.Set("login_challenge", hlc)
			registrationUrl.RawQuery = q.Encode()

			req, err := http.NewRequest("GET", registrationUrl.String(), nil)
			require.NoError(t, err)

			resp, err := c.browserClient.Do(req)
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusOK, resp.StatusCode)
			require.NoError(t, resp.Body.Close())

			// the registration page redirects us to the login page
			// which means we have a session now
			var oryCookie *http.Cookie
			currentURL, err := url.Parse(kratosPublicTS.URL)
			require.NoError(t, err)

			for _, c := range c.browserClient.Jar.Cookies(currentURL) {
				if c.Name == config.DefaultSessionCookieName {
					oryCookie = c
					break
				}
			}

			if oryCookie != nil {
				t.Log("[registrationTS] we expect to have been at the login screen and got an active flow. This means we have a session now")
				return
			}

			flowID := resp.Request.URL.Query().Get("flow")
			assert.NotEmpty(t, flowID)

			f := testhelpers.GetRegistrationFlow(t, c.browserClient, c.kratosPublicTS, flowID)
			require.NotNil(t, f)

			// continue the registration flow here
			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			values.Set("traits.foobar", c.identifier)
			values.Set("traits.username", c.identifier)
			values.Set("password", c.password)

			_, res := testhelpers.RegistrationMakeRequest(t, false, false, f, c.browserClient, values.Encode())
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
			return
		}

		if q.Has("flow") {
			c.callTrace = append(c.callTrace, RegistrationWithFlowID)
			t.Log("[registrationTS] registration flow is ignored here since it will be handled by the code above, we just need to return")
			return
		}
	})

	router.GET("/consent", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		t.Log("[consentTS] navigated to the consent ui")
		c := r.Context().Value(TestUIConfig).(*kratosUIConfig)
		c.callTrace = append(c.callTrace, Consent)

		q := r.URL.Query()
		consentChallenge := q.Get("consent_challenge")
		assert.NotEmpty(t, consentChallenge)

		if consentChallenge != "" {
			c.callTrace = append(c.callTrace, ConsentWithChallenge)
		}

		cr, resp, err := hydraAdminClient.GetOAuth2ConsentRequest(ctx).ConsentChallenge(q.Get("consent_challenge")).Execute()
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.ElementsMatch(t, cr.RequestedScope, []string{"profile", "email"})

		completedAcceptRequest, resp, err := hydraAdminClient.AcceptOAuth2ConsentRequest(context.Background()).AcceptOAuth2ConsentRequest(hydraclientgo.AcceptOAuth2ConsentRequest{
			Remember: &c.consentRemember,
		}).ConsentChallenge(q.Get("consent_challenge")).Execute()

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		if completedAcceptRequest != nil {
			c.callTrace = append(c.callTrace, ConsentAccept)
		}
		assert.NotNil(t, completedAcceptRequest)

		t.Logf("[consentTS] navigating to %s", completedAcceptRequest.RedirectTo)
		resp, err = c.browserClient.Get(completedAcceptRequest.RedirectTo)
		require.NoError(t, err)
		require.Equal(t, c.clientAppTS.URL, fmt.Sprintf("%s://%s", resp.Request.URL.Scheme, resp.Request.URL.Host))
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
		kc := ctx.Value(TestUIConfig).(*kratosUIConfig)
		kc.callTrace = append(kc.callTrace, CodeExchange)

		c.state.visits += 1
		t.Logf("[clientAppTS] handling a callback at client app %s", r.URL.String())
		if r.URL.Query().Has("code") {
			token, err := c.client.Exchange(r.Context(), r.URL.Query().Get("code"))
			require.NoError(t, err)

			if token != nil && token.AccessToken != "" {
				kc.callTrace = append(kc.callTrace, CodeExchangeWithToken)
			}
			c.state.tokens += 1
			t.Log("[clientAppTS] successfully exchanged code for token")
		} else {
			t.Error("[clientAppTS] code query parameter is missing")
		}
	})
	// A new OAuth client which will also function as the callback for the code exchange
	clientAppTS := testhelpers.NewHTTPTestServer(t, clientAppTSMiddleware)

	// we want to test if the registration ui is used if the flow contains an oauth2 login challenge
	// so we will have Hydra redirect to the base path of the test kratos ui server which
	// will then initiate the registration flow
	hydraAdmin, hydraPublic := newHydra(t, kratosUITS.URL+"/registration-ts", kratosUITS.URL+"/consent")

	hydraAdminClient = createHydraOAuth2ApiClient(hydraAdmin)
	clientID := createOAuth2Client(t, ctx, hydraAdminClient, []string{clientAppTS.URL}, "profile email", false)

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

	type state struct {
		cas clientAppState
	}
	doOAuthFlow := func(t *testing.T, ctx context.Context, oauthClient *oauth2.Config, browserClient *http.Client, expected state) {
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

		cac := ctx.Value(TestOAuthClientState).(*clientAppConfig)
		require.EqualValues(t, expected.cas, cac.state)
	}

	registerNewAccount := func(t *testing.T, ctx context.Context, browserClient *http.Client, identifier, password string) {
		// we need to create a new session directly with kratos
		f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, kratosPublicTS, false, false, false)
		require.NotNil(t, f)

		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("traits.foobar", identifier)
		values.Set("traits.username", identifier)
		values.Set("password", password)

		_, resp := testhelpers.RegistrationMakeRequest(t, false, false, f, browserClient, values.Encode())
		require.EqualValues(t, http.StatusOK, resp.StatusCode)

		var cookie *http.Cookie
		currentURL, err := url.Parse(kratosPublicTS.URL)
		require.NoError(t, err)

		for _, c := range browserClient.Jar.Cookies(currentURL) {
			if c.Name == config.DefaultSessionCookieName {
				cookie = c
				break
			}
		}

		require.NotNil(t, cookie, "expected exactly one session cookie to be set but got none")
	}

	t.Run("case=should accept oauth login request on registration", func(t *testing.T) {
		// this test initiates a new OAuth2 flow which goes directly to the registration page
		// we then create a new account through the registration flow
		// and expect the OAuth2 flow to succeed
		conf.MustSet(ctx, config.ViperKeySessionPersistentCookie, false)

		browserClient := testhelpers.NewClientWithCookieJar(t, nil, true)

		identifier := x.NewUUID().String()
		password := x.NewUUID().String()

		ct := make([]callTrace, 0)

		ctx = context.WithValue(ctx, TestUIConfig, &kratosUIConfig{
			identifier:       identifier,
			password:         password,
			browserClient:    browserClient,
			kratosPublicTS:   kratosPublicTS,
			clientAppTS:      clientAppTS,
			hydraAdminClient: hydraAdminClient,
			consentRemember:  true,
			callTrace:        ct,
		})

		ctx = context.WithValue(ctx, TestOAuthClientState, &clientAppConfig{
			client:      defaultClient,
			expectToken: true,
		})

		doOAuthFlow(t, ctx,
			defaultClient,
			browserClient,
			state{
				cas: clientAppState{
					visits: 1,
					tokens: 1,
				},
			})

		// we expect the following call trace
		expected := []callTrace{
			RegistrationUI,
			RegistrationWithOAuth2LoginChallenge,
			RegistrationWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		}
		require.ElementsMatchf(t, expected, ct, "expected the following call trace %+v but got %+v", expected, ct)
	})

	t.Run("case=should prompt the user for oauth login and consent even with session", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySessionPersistentCookie, true)

		clientID := createOAuth2Client(t, ctx, hydraAdminClient, []string{clientAppTS.URL}, "profile email", false)

		oauthClient := &oauth2.Config{
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

		browserClient := testhelpers.NewClientWithCookieJar(t, nil, true)
		identifier := x.NewUUID().String()
		password := x.NewUUID().String()

		registerNewAccount(t, ctx, browserClient, identifier, password)

		ct := make([]callTrace, 0)

		ctx = context.WithValue(ctx, TestUIConfig, &kratosUIConfig{
			identifier:       identifier,
			password:         password,
			browserClient:    browserClient,
			kratosPublicTS:   kratosPublicTS,
			clientAppTS:      clientAppTS,
			hydraAdminClient: hydraAdminClient,
			consentRemember:  true,
			callTrace:        ct,
		})

		ctx = context.WithValue(ctx, TestOAuthClientState, &clientAppConfig{
			client:      oauthClient,
			expectToken: true,
		})

		doOAuthFlow(t, ctx,
			oauthClient,
			browserClient,
			state{
				cas: clientAppState{
					visits: 1,
					tokens: 1,
				},
			})

		expected := []callTrace{
			RegistrationUI,
			RegistrationWithOAuth2LoginChallenge,
			LoginUI,
			LoginWithOAuth2LoginChallenge,
			LoginWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		}
		require.ElementsMatchf(t, expected, ct, "expected the following call trace %+v but got %+v", expected, ct)
	})

	t.Run("case=registration should redirect to login if skip=false and session exists", func(t *testing.T) {
		// we dont want to skip the consent page here
		clientSkipConsent := false
		consentRemember := false

		clientID := createOAuth2Client(t, ctx, hydraAdminClient, []string{clientAppTS.URL}, "profile email", clientSkipConsent)
		oauthClient := &oauth2.Config{
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

		browserClient := testhelpers.NewClientWithCookieJar(t, nil, true)
		identifier := x.NewUUID().String()
		password := x.NewUUID().String()

		ct := make([]callTrace, 0)

		kratosUIConfig := &kratosUIConfig{
			identifier:       identifier,
			password:         password,
			browserClient:    browserClient,
			kratosPublicTS:   kratosPublicTS,
			clientAppTS:      clientAppTS,
			hydraAdminClient: hydraAdminClient,
			consentRemember:  consentRemember,
			callTrace:        ct,
		}

		clientAppConfig := &clientAppConfig{
			client:      oauthClient,
			expectToken: true,
		}

		doSuccessfulOAuthFlow := func(t *testing.T) {
			t.Helper()

			ctx = context.WithValue(ctx, TestUIConfig, kratosUIConfig)

			ctx = context.WithValue(ctx, TestOAuthClientState, clientAppConfig)

			doOAuthFlow(t, ctx,
				oauthClient,
				browserClient,
				state{
					cas: clientAppState{
						visits: 1,
						tokens: 1,
					},
				})
		}

		doSuccessfulOAuthFlow(t)

		expected := []callTrace{
			RegistrationUI,
			RegistrationWithOAuth2LoginChallenge,
			RegistrationWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		}
		require.ElementsMatchf(t, expected, ct, "expected the following call trace %+v but got %+v", expected, ct)

		doOAuthFlow(t, ctx, oauthClient, browserClient, state{
			cas: clientAppState{
				visits: 1,
				tokens: 1,
			},
		})
		expected = append(expected,
			RegistrationUI,
			RegistrationWithOAuth2LoginChallenge,
			LoginUI,
			LoginWithOAuth2LoginChallenge,
			LoginWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		)
		require.ElementsMatchf(t, expected, ct, "expected the following call trace %+v but got %+v", expected, ct)
	})

	t.Run("case=should fail because the persistent Hydra session doesn't match the new Kratos session subject", func(t *testing.T) {
		// this test re-uses the previous test's clientID and oauthClient
		// but creates a new user account through the registration flow
		// since the session with the new user does not match the hydra session it should fail
		browserClient := testhelpers.NewClientWithCookieJar(t, nil, true)

		ct := make([]callTrace, 0)

		ctx = context.WithValue(ctx, TestUIConfig, &kratosUIConfig{
			identifier:       x.NewUUID().String(),
			password:         x.NewUUID().String(),
			kratosPublicTS:   kratosPublicTS,
			browserClient:    browserClient,
			clientAppTS:      clientAppTS,
			hydraAdminClient: hydraAdminClient,
			consentRemember:  false,
			callTrace:        ct,
		})

		ctx = context.WithValue(ctx, TestOAuthClientState, &clientAppConfig{
			client:      defaultClient,
			expectToken: false,
		})

		doOAuthFlow(t, ctx,
			defaultClient,
			browserClient,
			state{
				cas: clientAppState{
					visits: 0,
					tokens: 0,
				},
			})

		expected := []callTrace{
			RegistrationUI,
			RegistrationWithOAuth2LoginChallenge,
			RegistrationWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
		}
		require.ElementsMatchf(t, expected, ct, "expected the following call trace %+v but got %+v", expected, ct)
	})
}
