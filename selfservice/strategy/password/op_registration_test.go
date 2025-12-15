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
	"strings"
	"testing"

	"golang.org/x/oauth2"

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

func TestOAuth2ProviderRegistration(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, "selfservice.flows.registration.enable_legacy_one_step", true)

	kratosPublicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, x.NewRouterPublic(reg), x.NewRouterAdmin(reg))
	errTS := testhelpers.NewErrorTestServer(t, reg)
	redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

	var hydraAdminClient hydraclientgo.OAuth2API

	router := x.NewRouterPublic(reg)

	type contextKey string
	const (
		TestUIConfig         contextKey = "test-ui-config"
		TestOAuthClientState contextKey = "test-oauth-client-state"
	)

	router.HandleFunc("GET /login-ts", func(w http.ResponseWriter, r *http.Request) {
		t.Log("[loginTS] navigated to the login ui")
		c := r.Context().Value(TestUIConfig).(*testConfig)
		*c.callTrace = append(*c.callTrace, LoginUI)

		q := r.URL.Query()
		hlc := r.URL.Query().Get("login_challenge")

		if hlc != "" {
			*c.callTrace = append(*c.callTrace, LoginWithOAuth2LoginChallenge)
			return
		}

		if q.Has("flow") {
			*c.callTrace = append(*c.callTrace, LoginWithFlowID)
			lf := testhelpers.GetLoginFlow(t, c.browserClient, c.kratosPublicTS, q.Get("flow"))
			require.NotNil(t, lf)
			values := testhelpers.SDKFormFieldsToURLValues(lf.Ui.Nodes)
			values.Set("password", c.password)

			_, _ = testhelpers.LoginMakeRequest(t, false, false, lf, c.browserClient, values.Encode())
			t.Log("[loginTS] login flow is ignored here since it will be handled by the code above, we just need to return")
			return
		}
	})

	router.HandleFunc("GET /registration-ts", func(w http.ResponseWriter, r *http.Request) {
		t.Log("[registrationTS] navigated to the registration ui")
		c := r.Context().Value(TestUIConfig).(*testConfig)
		*c.callTrace = append(*c.callTrace, RegistrationUI)

		q := r.URL.Query()
		hlc := q.Get("login_challenge")

		if hlc != "" {
			*c.callTrace = append(*c.callTrace, RegistrationWithOAuth2LoginChallenge)
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

			// if the registration page redirects us to the login page
			// we will sign in which means we might have a session
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
			*c.callTrace = append(*c.callTrace, RegistrationWithFlowID)
			t.Log("[registrationTS] registration flow is ignored here since it will be handled by the code above, we just need to return")
			return
		}
	})

	router.HandleFunc("GET /consent", func(w http.ResponseWriter, r *http.Request) {
		t.Log("[consentTS] navigated to the consent ui")
		c := r.Context().Value(TestUIConfig).(*testConfig)
		*c.callTrace = append(*c.callTrace, Consent)

		q := r.URL.Query()
		consentChallenge := q.Get("consent_challenge")
		assert.NotEmpty(t, consentChallenge)

		if consentChallenge != "" {
			*c.callTrace = append(*c.callTrace, ConsentWithChallenge)
		}

		cr, resp, err := hydraAdminClient.GetOAuth2ConsentRequest(ctx).ConsentChallenge(q.Get("consent_challenge")).Execute()
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.ElementsMatch(t, cr.RequestedScope, c.requestedScope)

		if cr.GetSkip() {
			*c.callTrace = append(*c.callTrace, ConsentSkip)
		}

		if cr.Client.GetSkipConsent() {
			*c.callTrace = append(*c.callTrace, ConsentClientSkip)
		}

		completedAcceptRequest, resp, err := hydraAdminClient.AcceptOAuth2ConsentRequest(r.Context()).AcceptOAuth2ConsentRequest(hydraclientgo.AcceptOAuth2ConsentRequest{
			Remember:   &c.consentRemember,
			GrantScope: c.requestedScope,
		}).ConsentChallenge(q.Get("consent_challenge")).Execute()

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		require.NotNil(t, completedAcceptRequest)
		*c.callTrace = append(*c.callTrace, ConsentAccept)

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
		kc := ctx.Value(TestUIConfig).(*testConfig)
		*kc.callTrace = append(*kc.callTrace, CodeExchange)

		c.state.visits += 1
		t.Logf("[clientAppTS] handling a callback at client app %s", r.URL.String())
		if r.URL.Query().Has("code") {
			token, err := c.client.Exchange(r.Context(), r.URL.Query().Get("code"))
			require.NoError(t, err)

			if token != nil && token.AccessToken != "" {
				t.Log("[clientAppTS] successfully exchanged code for token")
				*kc.callTrace = append(*kc.callTrace, CodeExchangeWithToken)
				c.state.tokens += 1
			} else {
				t.Log("[clientAppTS] did not receive a token")
			}
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

	doOAuthFlow := func(t *testing.T, ctx context.Context, oauthClient *oauth2.Config, browserClient *http.Client) {
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

	registerNewAccount := func(t *testing.T, browserClient *http.Client, identifier, password string) {
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

	// important, we will set the persistent cookie to false for most of the tests here
	// once this is true, the behavior of the consent flow changes
	conf.MustSet(ctx, config.ViperKeySessionPersistentCookie, false)

	t.Run("case=should accept oauth login request on registration", func(t *testing.T) {
		// this test initiates a new OAuth2 flow which goes directly to the registration page
		// we then create a new account through the registration flow
		// and expect the OAuth2 flow to succeed
		scopes := []string{"profile", "email"}
		clientID := createOAuth2Client(t, ctx, hydraAdminClient, []string{clientAppTS.URL}, strings.Join(scopes, " "), false)

		oauth2Client := &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: "client-secret",
			Endpoint: oauth2.Endpoint{
				AuthURL:   hydraPublic + "/oauth2/auth",
				TokenURL:  hydraPublic + "/oauth2/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
			Scopes:      scopes,
			RedirectURL: clientAppTS.URL,
		}
		browserClient := testhelpers.NewClientWithCookieJar(t, nil, nil)

		identifier := x.NewUUID().String()
		password := x.NewUUID().String()

		ct := make([]callTrace, 0)

		ctx = context.WithValue(ctx, TestUIConfig, &testConfig{
			identifier:       identifier,
			password:         password,
			browserClient:    browserClient,
			kratosPublicTS:   kratosPublicTS,
			clientAppTS:      clientAppTS,
			hydraAdminClient: hydraAdminClient,
			consentRemember:  true,
			callTrace:        &ct,
			requestedScope:   scopes,
		})

		clientAS := clientAppState{
			visits: 0,
			tokens: 0,
		}

		ctx = context.WithValue(ctx, TestOAuthClientState, &clientAppConfig{
			client:      oauth2Client,
			expectToken: true,
			state:       &clientAS,
		})

		doOAuthFlow(t, ctx,
			oauth2Client,
			browserClient)

		assert.EqualValues(t, clientAppState{
			visits: 1,
			tokens: 1,
		}, clientAS)

		expected := []callTrace{
			RegistrationUI,
			RegistrationWithOAuth2LoginChallenge,
			RegistrationUI,
			RegistrationWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		}
		require.ElementsMatch(t, expected, ct, "expected the call trace to match")
	})

	t.Run("case=registration with session should redirect to login to re-authenticate and to consent", func(t *testing.T) {
		// this test registers a new account which sets a session
		// we then initiate a new OAuth2 flow which should redirect us to the registration page
		// the registration page does a session validation and retrieves the loginRequest from Hydra
		// which in this case will indicate we cannot skip the login flow (since the there is no previous OAuth flow associated)
		// we then get redirected to the the login page with refresh=true
		// we then sign in and expect to be redirected to the consent page
		// and then back to the client app
		scopes := []string{"profile", "email", "offline_access"}

		clientID := createOAuth2Client(t, ctx, hydraAdminClient, []string{clientAppTS.URL}, strings.Join(scopes, " "), false)

		oauthClient := &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: "client-secret",
			Endpoint: oauth2.Endpoint{
				AuthURL:   hydraPublic + "/oauth2/auth",
				TokenURL:  hydraPublic + "/oauth2/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
			Scopes:      scopes,
			RedirectURL: clientAppTS.URL,
		}

		browserClient := testhelpers.NewClientWithCookieJar(t, nil, nil)
		identifier := x.NewUUID().String()
		password := x.NewUUID().String()

		ct := make([]callTrace, 0)

		tc := &testConfig{
			identifier:       identifier,
			password:         password,
			browserClient:    browserClient,
			kratosPublicTS:   kratosPublicTS,
			clientAppTS:      clientAppTS,
			hydraAdminClient: hydraAdminClient,
			consentRemember:  true,
			requestedScope:   scopes,
			callTrace:        &ct,
		}

		ctx = context.WithValue(ctx, TestUIConfig, tc)

		clientAS := clientAppState{
			visits: 0,
			tokens: 0,
		}

		ctx = context.WithValue(ctx, TestOAuthClientState, &clientAppConfig{
			client:      oauthClient,
			expectToken: true,
			state:       &clientAS,
		})

		registerNewAccount(t, browserClient, identifier, password)

		require.ElementsMatch(t, []callTrace{
			RegistrationUI,
			RegistrationWithFlowID,
		}, ct, "expected the call trace to match")

		// reset the call trace
		ct = []callTrace{}
		tc.callTrace = &ct

		doOAuthFlow(t, ctx, oauthClient, browserClient)

		assert.EqualValues(t, clientAppState{
			visits: 1,
			tokens: 1,
		}, clientAS)

		expected := []callTrace{
			RegistrationUI,
			RegistrationWithOAuth2LoginChallenge,
			LoginUI,
			LoginWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		}
		require.ElementsMatch(t, expected, ct, "expected the call trace to match")
	})

	t.Run("case=registration should redirect to login if session exists and skip=false", func(t *testing.T) {
		// we dont want to skip the consent page here
		// we want the registration page to redirect to the login page
		// since we have a session but do not skip the consent page
		clientSkipConsent := false
		consentRemember := false

		scopes := []string{"profile", "email", "offline_access"}

		clientID := createOAuth2Client(t, ctx, hydraAdminClient, []string{clientAppTS.URL}, strings.Join(scopes, " "), clientSkipConsent)
		oauthClient := &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: "client-secret",
			Endpoint: oauth2.Endpoint{
				AuthURL:   hydraPublic + "/oauth2/auth",
				TokenURL:  hydraPublic + "/oauth2/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
			Scopes:      scopes,
			RedirectURL: clientAppTS.URL,
		}

		browserClient := testhelpers.NewClientWithCookieJar(t, nil, nil)
		identifier := x.NewUUID().String()
		password := x.NewUUID().String()

		ct := make([]callTrace, 0)

		tc := &testConfig{
			identifier:       identifier,
			password:         password,
			browserClient:    browserClient,
			kratosPublicTS:   kratosPublicTS,
			clientAppTS:      clientAppTS,
			hydraAdminClient: hydraAdminClient,
			consentRemember:  consentRemember,
			requestedScope:   scopes,
			callTrace:        &ct,
		}

		clientAppConfig := &clientAppConfig{
			client:      oauthClient,
			expectToken: true,
		}

		expected := []callTrace{
			RegistrationUI,
			RegistrationWithOAuth2LoginChallenge,
			RegistrationUI,
			RegistrationWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		}

		// set the global context values
		ctx = context.WithValue(ctx, TestUIConfig, tc)
		ctx = context.WithValue(ctx, TestOAuthClientState, clientAppConfig)

		doSuccessfulOAuthFlow := func(t *testing.T) {
			t.Helper()

			clientAS := clientAppState{
				visits: 0,
				tokens: 0,
			}
			clientAppConfig.state = &clientAS

			doOAuthFlow(t, ctx,
				oauthClient,
				browserClient)
			assert.EqualValues(t, clientAppState{
				visits: 1,
				tokens: 1,
			}, clientAS)

			require.ElementsMatchf(t, expected, ct, "expected the call trace to match")
		}

		doSuccessfulOAuthFlow(t)

		// reset our state on the client app
		clientAS := clientAppState{
			visits: 0,
			tokens: 0,
		}

		clientAppConfig.state = &clientAS

		// we should now have a session, but not skip the consent page
		doOAuthFlow(t, ctx, oauthClient, browserClient)
		assert.EqualValues(t, clientAppState{
			visits: 1,
			tokens: 1,
		}, clientAS)

		expected = append(expected,
			RegistrationUI,
			RegistrationWithOAuth2LoginChallenge,
			LoginUI,
			LoginWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		)
		require.ElementsMatchf(t, expected, ct, "expected the call trace to match")
	})

	t.Run("case=consent should be skipped if client is configured to skip", func(t *testing.T) {
		clientSkipConsent := true
		scopes := []string{"profile", "email", "offline_access"}
		clientID := createOAuth2Client(t, ctx, hydraAdminClient, []string{clientAppTS.URL}, strings.Join(scopes, " "), clientSkipConsent)

		oauthClient := &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: "client-secret",
			Endpoint: oauth2.Endpoint{
				AuthURL:   hydraPublic + "/oauth2/auth",
				TokenURL:  hydraPublic + "/oauth2/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
			Scopes:      scopes,
			RedirectURL: clientAppTS.URL,
		}

		browserClient := testhelpers.NewClientWithCookieJar(t, nil, nil)
		identifier := x.NewUUID().String()
		password := x.NewUUID().String()

		ct := make([]callTrace, 0)

		tc := &testConfig{
			identifier:       identifier,
			password:         password,
			browserClient:    browserClient,
			kratosPublicTS:   kratosPublicTS,
			clientAppTS:      clientAppTS,
			hydraAdminClient: hydraAdminClient,
			consentRemember:  false,
			requestedScope:   scopes,
			callTrace:        &ct,
		}

		clientAppConfig := &clientAppConfig{
			client:      oauthClient,
			expectToken: true,
		}

		expected := []callTrace{
			RegistrationUI,
			RegistrationWithOAuth2LoginChallenge,
			RegistrationUI,
			RegistrationWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentClientSkip,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		}

		// set the global context values
		ctx = context.WithValue(ctx, TestUIConfig, tc)
		ctx = context.WithValue(ctx, TestOAuthClientState, clientAppConfig)

		clientAS := clientAppState{
			visits: 0,
			tokens: 0,
		}
		clientAppConfig.state = &clientAS

		doOAuthFlow(t, ctx,
			oauthClient,
			browserClient)
		assert.EqualValues(t, clientAppState{
			visits: 1,
			tokens: 1,
		}, clientAS)

		require.ElementsMatchf(t, expected, ct, "expected the call trace to match")
	})

	t.Run("case=consent should be skipped if user has a session and has already consented", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySessionPersistentCookie, true)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySessionPersistentCookie, false)
		})

		consentRemember := true

		scopes := []string{"profile", "email", "offline_access"}

		clientID := createOAuth2Client(t, ctx, hydraAdminClient, []string{clientAppTS.URL}, strings.Join(scopes, " "), false)
		oauthClient := &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: "client-secret",
			Endpoint: oauth2.Endpoint{
				AuthURL:   hydraPublic + "/oauth2/auth",
				TokenURL:  hydraPublic + "/oauth2/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
			Scopes:      scopes,
			RedirectURL: clientAppTS.URL,
		}

		browserClient := testhelpers.NewClientWithCookieJar(t, nil, nil)
		identifier := x.NewUUID().String()
		password := x.NewUUID().String()

		ct := make([]callTrace, 0)

		kratosUIConfig := &testConfig{
			identifier:       identifier,
			password:         password,
			browserClient:    browserClient,
			kratosPublicTS:   kratosPublicTS,
			clientAppTS:      clientAppTS,
			hydraAdminClient: hydraAdminClient,
			consentRemember:  consentRemember,
			requestedScope:   scopes,
			callTrace:        &ct,
		}

		clientAppConfig := &clientAppConfig{
			client:      oauthClient,
			expectToken: true,
		}

		expected := []callTrace{
			RegistrationUI,
			RegistrationWithOAuth2LoginChallenge,
			RegistrationUI,
			RegistrationWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		}

		// set the global context values
		ctx = context.WithValue(ctx, TestUIConfig, kratosUIConfig)
		ctx = context.WithValue(ctx, TestOAuthClientState, clientAppConfig)

		doSuccessfulOAuthFlow := func(t *testing.T) {
			t.Helper()

			clientAS := clientAppState{
				visits: 0,
				tokens: 0,
			}
			clientAppConfig.state = &clientAS

			doOAuthFlow(t, ctx,
				oauthClient,
				browserClient)
			assert.EqualValues(t, clientAppState{
				visits: 1,
				tokens: 1,
			}, clientAS)

			require.ElementsMatchf(t, expected, ct, "expected the call trace to match")
		}

		doSuccessfulOAuthFlow(t)

		// reset our state on the client app
		clientAS := clientAppState{
			visits: 0,
			tokens: 0,
		}

		clientAppConfig.state = &clientAS

		// reset the call trace
		ct = []callTrace{}
		kratosUIConfig.callTrace = &ct

		// we should now have a session, but not skip the consent page
		doOAuthFlow(t, ctx, oauthClient, browserClient)
		assert.EqualValues(t, clientAppState{
			visits: 1,
			tokens: 1,
		}, clientAS)

		expected = []callTrace{
			RegistrationUI,
			RegistrationWithOAuth2LoginChallenge,
			Consent,
			ConsentWithChallenge,
			ConsentSkip,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		}
		require.ElementsMatchf(t, expected, ct, "expected the call trace to match")
	})

	t.Run("case=should fail because the persistent Hydra session doesn't match the new Kratos session subject", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySessionPersistentCookie, true)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySessionPersistentCookie, false)
		})

		// this test re-uses the previous oauthClient
		// but creates a new user account through the registration flow
		// since the session with the new user does not match the hydra session it should fail
		scopes := []string{"profile", "email", "offline_access"}
		clientID := createOAuth2Client(t, ctx, hydraAdminClient, []string{clientAppTS.URL}, strings.Join(scopes, " "), false)

		oauthClient := &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: "client-secret",
			Endpoint: oauth2.Endpoint{
				AuthURL:   hydraPublic + "/oauth2/auth",
				TokenURL:  hydraPublic + "/oauth2/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
			Scopes:      scopes,
			RedirectURL: clientAppTS.URL,
		}

		browserClient := testhelpers.NewClientWithCookieJar(t, nil, nil)

		ct := make([]callTrace, 0)

		kratosUIConfig := &testConfig{
			identifier:       x.NewUUID().String(),
			password:         x.NewUUID().String(),
			kratosPublicTS:   kratosPublicTS,
			browserClient:    browserClient,
			clientAppTS:      clientAppTS,
			hydraAdminClient: hydraAdminClient,
			consentRemember:  true,
			callTrace:        &ct,
			requestedScope:   scopes,
		}

		ctx = context.WithValue(ctx, TestUIConfig, kratosUIConfig)

		clientAppConfig := &clientAppConfig{
			client:      oauthClient,
			expectToken: false,
		}

		ctx = context.WithValue(ctx, TestOAuthClientState, clientAppConfig)

		expected := []callTrace{
			RegistrationUI,
			RegistrationWithOAuth2LoginChallenge,
			RegistrationUI,
			RegistrationWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		}

		doSuccessfulOAuthFlow := func(t *testing.T) {
			t.Helper()

			clientAS := clientAppState{
				visits: 0,
				tokens: 0,
			}
			clientAppConfig.state = &clientAS

			doOAuthFlow(t, ctx,
				oauthClient,
				browserClient)

			assert.EqualValues(t, clientAppState{
				visits: 1,
				tokens: 1,
			}, clientAS)

			require.ElementsMatch(t, expected, ct, "expected the call trace to match")
		}

		doSuccessfulOAuthFlow(t)

		clientAS := clientAppState{
			visits: 0,
			tokens: 0,
		}
		clientAppConfig.state = &clientAS

		currentURL, err := url.Parse(kratosPublicTS.URL)
		require.NoError(t, err)
		cookies := browserClient.Jar.Cookies(currentURL)

		// remove the kratos session so we can register a new account
		for _, c := range cookies {
			if c.Name == config.DefaultSessionCookieName {
				c.MaxAge = -1
				c.Value = ""
				break
			}
		}
		browserClient.Jar.SetCookies(currentURL, cookies)

		kratosUIConfig.identifier = x.NewUUID().String()
		kratosUIConfig.password = x.NewUUID().String()

		// reset the call trace
		ct = []callTrace{}
		kratosUIConfig.callTrace = &ct

		doOAuthFlow(t, ctx,
			oauthClient,
			browserClient)

		assert.EqualValues(t, clientAppState{
			visits: 0,
			tokens: 0,
		}, clientAS)

		expected = []callTrace{
			RegistrationUI,
			RegistrationWithOAuth2LoginChallenge,
			RegistrationUI,
			RegistrationWithFlowID,
		}
		require.ElementsMatch(t, expected, ct, "expected the call trace to match")
	})
}
