// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password_test

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/tidwall/gjson"
	"github.com/urfave/negroni"
	"golang.org/x/oauth2"

	"github.com/gofrs/uuid"

	hydraclientgo "github.com/ory/hydra-client-go/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/x"
)

func TestOAuth2Provider(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(
		ctx,
		config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword),
		map[string]interface{}{"enabled": true},
	)

	var hydraAdminClient hydraclientgo.OAuth2Api

	identifier, pwd := x.NewUUID().String(), "password"

	var testRequireLogin atomic.Bool
	testRequireLogin.Store(true)

	router := x.NewRouterPublic()
	kratosPublicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())
	errTS := testhelpers.NewErrorTestServer(t, reg)
	redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

	router.GET("/login-ts", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		t.Log("[loginTS] navigated to the login ui")
		c := r.Context().Value(TestUIConfig).(*testConfig)
		*c.callTrace = append(*c.callTrace, LoginUI)

		q := r.URL.Query()
		hlc := r.URL.Query().Get("login_challenge")
		if hlc != "" {
			*c.callTrace = append(*c.callTrace, LoginWithOAuth2LoginChallenge)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, c.browserClient, kratosPublicTS, false, false, false, !testRequireLogin.Load(), testhelpers.InitFlowWithOAuth2LoginChallenge(hlc))
			require.NotNil(t, f)

			values := url.Values{"method": {"password"}, "identifier": {identifier}, "password": {pwd}, "csrf_token": {x.FakeCSRFToken}}.Encode()
			_, res := testhelpers.LoginMakeRequest(t, false, false, f, c.browserClient, values)

			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		}

		if q.Has("flow") {
			*c.callTrace = append(*c.callTrace, LoginWithFlowID)
			t.Log("[loginTS] login flow is ignored here since it will be handled by the code above, we just need to return")
			return
		}
	})

	router.GET("/consent", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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

		if completedAcceptRequest != nil {
			*c.callTrace = append(*c.callTrace, ConsentAccept)
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

	conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})
	conf.MustSet(ctx, config.ViperKeySelfServiceErrorUI, errTS.URL+"/error-ts")
	conf.MustSet(ctx, config.ViperKeySelfServiceLoginUI, kratosUITS.URL+"/login-ts")
	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, redirTS.URL+"/return-ts")
	conf.MustSet(ctx, config.ViperKeySessionPersistentCookie, true)

	testhelpers.SetDefaultIdentitySchemaFromRaw(conf, loginSchema)
	createIdentity(ctx, reg, t, identifier, pwd)

	hydraAdmin, hydraPublic := newHydra(t, kratosUITS.URL+"/login", kratosUITS.URL+"/consent")
	conf.MustSet(ctx, config.ViperKeyOAuth2ProviderURL, hydraAdmin)

	hydraAdminClient = createHydraOAuth2ApiClient(hydraAdmin)

	loginToAccount := func(t *testing.T, browserClient *http.Client) {
		f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, kratosPublicTS, false, false, false, false)

		values := url.Values{"method": {"password"}, "identifier": {identifier}, "password": {pwd}, "csrf_token": {x.FakeCSRFToken}}.Encode()

		body, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, values)

		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, identifier, gjson.Get(body, "identity.traits.subject").String(), "%s", body)
	}

	t.Run("should prompt the user for login and consent", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySessionPersistentCookie, false)

		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySessionPersistentCookie, true)
		})
		browserClient := testhelpers.NewClientWithCookieJar(t, nil, true)

		scopes := []string{"profile", "email"}
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

		clientAC := &clientAppConfig{
			client: oauthClient,
			state: &clientAppState{
				visits: 0,
				tokens: 0,
			},
			expectToken: true,
		}

		ctx = context.WithValue(ctx, TestOAuthClientState, clientAC)

		tc := &testConfig{
			browserClient:   browserClient,
			kratosPublicTS:  kratosPublicTS,
			clientAppTS:     clientAppTS,
			callTrace:       new([]callTrace),
			requestedScope:  scopes,
			consentRemember: true,
		}
		ctx = context.WithValue(ctx, TestUIConfig, tc)

		doOAuthFlow(t, ctx, oauthClient, browserClient)

		assert.EqualValues(t, clientAppState{
			visits: 1,
			tokens: 1,
		}, clientAC.state)

		expected := []callTrace{
			LoginUI,
			LoginWithOAuth2LoginChallenge,
			LoginUI,
			LoginWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		}
		require.ElementsMatch(t, expected, tc)
	})

	t.Run("should prompt the user for login and consent again", func(t *testing.T) {
		browserClient := testhelpers.NewClientWithCookieJar(t, nil, true)

		scopes := []string{"profile", "email"}
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

		clientAC := &clientAppConfig{
			client: oauthClient,
			state: &clientAppState{
				visits: 0,
				tokens: 0,
			},
			expectToken: true,
		}

		ctx = context.WithValue(ctx, TestOAuthClientState, clientAC)

		tc := &testConfig{
			browserClient:   browserClient,
			kratosPublicTS:  kratosPublicTS,
			clientAppTS:     clientAppTS,
			callTrace:       new([]callTrace),
			requestedScope:  scopes,
			consentRemember: false,
		}
		ctx = context.WithValue(ctx, TestUIConfig, tc)

		doOAuthFlow(t, ctx, oauthClient, browserClient)

		assert.EqualValues(t, clientAppState{
			visits: 1,
			tokens: 1,
		}, clientAC.state)

		expected := []callTrace{
			LoginUI,
			LoginWithOAuth2LoginChallenge,
			LoginUI,
			LoginWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		}
		require.ElementsMatch(t, expected, tc)

		tc.callTrace = new([]callTrace)
		doOAuthFlow(t, ctx, oauthClient, browserClient)

		assert.EqualValues(t, clientAppState{
			visits: 2,
			tokens: 2,
		}, clientAC.state)

		expected = []callTrace{
			LoginUI,
			LoginWithOAuth2LoginChallenge,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		}
		require.ElementsMatch(t, expected, tc)
	})

	t.Run("should prompt the user for consent, but not for login", func(t *testing.T) {
		browserClient := testhelpers.NewClientWithCookieJar(t, nil, true)

		loginToAccount(t, browserClient)

		scopes := []string{"profile", "email"}
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

		clientAC := &clientAppConfig{
			client: oauthClient,
			state: &clientAppState{
				visits: 0,
				tokens: 0,
			},
			expectToken: true,
		}

		ctx = context.WithValue(ctx, TestOAuthClientState, clientAC)

		tc := &testConfig{
			browserClient:   browserClient,
			kratosPublicTS:  kratosPublicTS,
			clientAppTS:     clientAppTS,
			callTrace:       new([]callTrace),
			requestedScope:  scopes,
			consentRemember: false,
		}
		ctx = context.WithValue(ctx, TestUIConfig, tc)

		doOAuthFlow(t, ctx, oauthClient, browserClient)

		assert.EqualValues(t, clientAppState{
			visits: 1,
			tokens: 1,
		}, clientAC.state)

		expected := []callTrace{
			LoginUI,
			LoginWithOAuth2LoginChallenge,
			LoginUI,
			LoginWithFlowID,
			Consent,
			ConsentWithChallenge,
			ConsentAccept,
			CodeExchange,
			CodeExchangeWithToken,
		}
		require.ElementsMatch(t, expected, tc)
	})

	t.Run("should fail when Hydra session subject doesn't match the subject authenticated by Kratos", func(t *testing.T) {
		t.Skip()
		reg.WithHydra(&AcceptWrongSubject{h: reg.Hydra().(*hydra.DefaultHydra)})
		browserClient := testhelpers.NewClientWithCookieJar(t, nil, true)

		scopes := []string{"profile", "email"}
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
		tc := &testConfig{
			browserClient:  browserClient,
			kratosPublicTS: kratosPublicTS,
			clientAppTS:    clientAppTS,
			callTrace:      new([]callTrace),
		}

		ctx = context.WithValue(ctx, TestUIConfig, tc)

		doOAuthFlow(t, ctx, oauthClient, browserClient)

		expected := []callTrace{
			LoginUI,
			LoginWithOAuth2LoginChallenge,
			LoginUI,
			LoginWithFlowID,
		}
		require.ElementsMatch(t, expected, ctx.Value(TestUIConfig).(*testConfig))
	})
}

var _ hydra.Hydra = &AcceptWrongSubject{}

type AcceptWrongSubject struct {
	h *hydra.DefaultHydra
}

func (h *AcceptWrongSubject) AcceptLoginRequest(ctx context.Context, params hydra.AcceptLoginRequestParams) (string, error) {
	params.IdentityID = uuid.Must(uuid.NewV4()).String()
	return h.h.AcceptLoginRequest(ctx, params)
}

func (h *AcceptWrongSubject) GetLoginRequest(ctx context.Context, loginChallenge string) (*hydraclientgo.OAuth2LoginRequest, error) {
	return h.h.GetLoginRequest(ctx, loginChallenge)
}
