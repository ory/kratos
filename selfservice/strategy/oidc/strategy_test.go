package oidc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/sqlxx"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/x"
)

const debugRedirects = false

func TestStrategy(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var (
		conf, reg = internal.NewFastRegistryWithMocks(t)
		subject   string
		scope     []string
	)

	remoteAdmin, remotePublic, hydraIntegrationTSURL := newHydra(t, &subject, &scope)
	returnTS := newReturnTs(t, reg)
	uiTS := newUI(t, reg)
	errTS := testhelpers.NewErrorTestServer(t, reg)
	ts, tsA := testhelpers.NewKratosServers(t)

	viperSetProviderConfig(
		t,
		conf,
		newOIDCProvider(t, ts, remotePublic, remoteAdmin, "valid", "client"),
		oidc.Configuration{
			Provider:     "generic",
			ID:           "invalid-issuer",
			ClientID:     "client",
			ClientSecret: "secret",
			IssuerURL:    strings.Replace(remotePublic, "127.0.0.1", "localhost", 1) + "/",
			Mapper:       "file://./stub/oidc.hydra.jsonnet",
		},
	)
	testhelpers.InitKratosServers(t, reg, ts, tsA)
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
	conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter,
		identity.CredentialsTypeOIDC.String()), []config.SelfServiceHook{{Name: "session"}})

	t.Logf("Kratos Public URL: %s", ts.URL)
	t.Logf("Kratos Error URL: %s", errTS.URL)
	t.Logf("Hydra Public URL: %s", remotePublic)
	t.Logf("Hydra Admin URL: %s", remoteAdmin)
	t.Logf("Hydra Integration URL: %s", hydraIntegrationTSURL)
	t.Logf("Return URL: %s", returnTS.URL)

	subject = "foo@bar.com"
	scope = []string{}

	// assert form values
	var afv = func(t *testing.T, flowID uuid.UUID, provider string) (action string) {
		var config *form.HTMLForm
		if req, err := reg.RegistrationFlowPersister().GetRegistrationFlow(context.Background(), flowID); err == nil {
			require.EqualValues(t, req.ID, flowID)
			method := req.Methods[identity.CredentialsTypeOIDC]
			require.NotNil(t, method)
			config = method.Config.FlowMethodConfigurator.(*form.HTMLForm)
			require.NotNil(t, config)
		} else if req, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), flowID); err == nil {
			require.EqualValues(t, req.ID, flowID)
			method := req.Methods[identity.CredentialsTypeOIDC]
			require.NotNil(t, method)
			config = method.Config.FlowMethodConfigurator.(*form.HTMLForm)
			require.NotNil(t, config)
		} else {
			require.NoError(t, err)
			return
		}

		assert.Equal(t, "POST", config.Method)

		var found bool
		for _, field := range config.Fields {
			if field.Name == "provider" && field.Value == provider {
				found = true
				break
			}
		}
		require.True(t, found)

		return config.Action
	}

	var requestAction = func(flowID uuid.UUID) string {
		return ts.URL + oidc.RouteBase + "/auth/" + flowID.String()
	}

	var makeRequestWithCookieJar = func(t *testing.T, provider string, action string, fv url.Values, jar *cookiejar.Jar) (*http.Response, []byte) {
		fv.Set("provider", provider)
		res, err := newClient(t, jar).PostForm(action, fv)
		require.NoError(t, err)

		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		require.NoError(t, err)

		require.Equal(t, 200, res.StatusCode, "%s: %s\n\t%s", action, res.Request.URL.String(), body)

		return res, body
	}

	var makeRequest = func(t *testing.T, provider string, action string, fv url.Values) (*http.Response, []byte) {
		return makeRequestWithCookieJar(t, provider, action, fv, nil)
	}

	var assertSystemError = func(t *testing.T, res *http.Response, body []byte, code int, reason string) {
		require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)

		assert.Equal(t, int64(code), gjson.GetBytes(body, "0.code").Int(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), reason, "%s", body)
	}

	// assert system error (redirect to error endpoint)
	var asem = func(t *testing.T, res *http.Response, body []byte, code int, reason string) {
		require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)

		assert.Equal(t, int64(code), gjson.GetBytes(body, "0.code").Int(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "0.message").String(), reason, "%s", body)
	}

	// assert ui error (redirect to login/registration ui endpoint)
	var aue = func(t *testing.T, res *http.Response, body []byte, reason string) {
		require.Contains(t, res.Request.URL.String(), uiTS.URL, "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "methods.oidc.config.messages.0.text").String(), reason, "%s", body)
	}

	// assert identity (success)
	var ai = func(t *testing.T, res *http.Response, body []byte) {
		assert.Contains(t, res.Request.URL.String(), returnTS.URL)
		assert.Equal(t, subject, gjson.GetBytes(body, "identity.traits.subject").String(), "%s", body)
	}

	var newLoginFlow = func(t *testing.T, redirectTo string, exp time.Duration) (req *login.Flow) {
		// Use NewLoginFlow to instantiate the request but change the things we need to control a copy of it.
		req, err := reg.LoginHandler().NewLoginFlow(httptest.NewRecorder(),
			&http.Request{URL: urlx.ParseOrPanic(redirectTo)}, flow.TypeBrowser)
		require.NoError(t, err)
		req.RequestURL = redirectTo
		req.ExpiresAt = time.Now().Add(exp)
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), req))

		// sanity check
		got, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), req.ID)
		require.NoError(t, err)
		require.Len(t, got.Methods, len(req.Methods), "%+v", got)

		return
	}

	var newRegistrationFlow = func(t *testing.T, redirectTo string, exp time.Duration) *registration.Flow {
		// Use NewLoginFlow to instantiate the request but change the things we need to control a copy of it.
		req, err := reg.RegistrationHandler().NewRegistrationFlow(httptest.NewRecorder(),
			&http.Request{URL: urlx.ParseOrPanic(redirectTo)}, flow.TypeBrowser)
		require.NoError(t, err)
		req.RequestURL = redirectTo
		req.ExpiresAt = time.Now().Add(exp)
		require.NoError(t, reg.RegistrationFlowPersister().UpdateRegistrationFlow(context.Background(), req))

		// sanity check
		got, err := reg.RegistrationFlowPersister().GetRegistrationFlow(context.Background(), req.ID)
		require.NoError(t, err)
		require.Len(t, got.Methods, len(req.Methods), "%+v", req)

		return req
	}

	t.Run("case=should fail because provider does not exist", func(t *testing.T) {
		doesNotExist := x.NewUUID()
		res, body := makeRequest(t, "provider-does-not-exist", requestAction(doesNotExist), url.Values{})
		assertSystemError(t, res, body, http.StatusNotFound, "is unknown or has not been configured")
	})

	t.Run("case=should fail because the issuer is mismatching", func(t *testing.T) {
		doesNotExist := x.NewUUID()
		res, body := makeRequest(t, "invalid-issuer", requestAction(doesNotExist), url.Values{})
		assertSystemError(t, res, body, http.StatusInternalServerError, "issuer did not match the issuer returned by provider")
	})

	t.Run("case=should fail because flow does not exist", func(t *testing.T) {
		doesNotExist := x.NewUUID()
		res, body := makeRequest(t, "valid", requestAction(doesNotExist), url.Values{})
		asem(t, res, body, http.StatusNotFound, "Unable to locate the resource")
	})

	t.Run("case=should fail because the login flow is expired", func(t *testing.T) {
		r := newLoginFlow(t, returnTS.URL, -time.Minute)
		action := afv(t, r.ID, "valid")
		t.Logf("action: %s id: %s", action, r.ID)
		res, body := makeRequest(t, "valid", action, url.Values{})

		assert.NotEqual(t, r.ID, gjson.GetBytes(body, "id"))
		require.Contains(t, res.Request.URL.String(), uiTS.URL, "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "messages.0.text").String(), "flow expired", "%s", body)
	})

	t.Run("case=should fail because the registration flow is expired", func(t *testing.T) {
		r := newRegistrationFlow(t, returnTS.URL, -time.Minute)
		action := afv(t, r.ID, "valid")
		res, body := makeRequest(t, "valid", action, url.Values{})

		assert.NotEqual(t, r.ID, gjson.GetBytes(body, "id"))
		require.Contains(t, res.Request.URL.String(), uiTS.URL, "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "messages.0.text").String(), "flow expired", "%s", body)
	})

	t.Run("case=should fail registration because scope was not provided", func(t *testing.T) {
		subject = "foo@bar.com"
		scope = []string{}

		r := newRegistrationFlow(t, returnTS.URL, time.Minute)
		action := afv(t, r.ID, "valid")
		res, body := makeRequest(t, "valid", action, url.Values{})
		aue(t, res, body, "no id_token was returned")
	})

	t.Run("case=should fail login because scope was not provided", func(t *testing.T) {
		r := newLoginFlow(t, returnTS.URL, time.Minute)
		action := afv(t, r.ID, "valid")
		res, body := makeRequest(t, "valid", action, url.Values{})
		aue(t, res, body, "no id_token was returned")
	})

	t.Run("case=should fail registration flow because subject is not an email", func(t *testing.T) {
		subject = "not-an-email"
		scope = []string{"openid"}

		r := newRegistrationFlow(t, returnTS.URL, time.Minute)
		action := afv(t, r.ID, "valid")
		res, body := makeRequest(t, "valid", action, url.Values{})

		require.Contains(t, res.Request.URL.String(), uiTS.URL, "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "methods.oidc.config.fields.#(name==traits.subject).messages.0").String(), "is not valid", "%s", body)
	})

	t.Run("case=register and then login", func(t *testing.T) {
		subject = "register-then-login@ory.sh"
		scope = []string{"openid"}

		t.Run("case=should pass registration", func(t *testing.T) {
			r := newRegistrationFlow(t, returnTS.URL, time.Minute)
			action := afv(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			ai(t, res, body)
		})

		t.Run("case=should pass login", func(t *testing.T) {
			r := newLoginFlow(t, returnTS.URL, time.Minute)
			action := afv(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			ai(t, res, body)
		})
	})

	t.Run("case=login without registered account", func(t *testing.T) {
		subject = "login-without-register@ory.sh"
		scope = []string{"openid"}

		t.Run("case=should pass login", func(t *testing.T) {
			r := newLoginFlow(t, returnTS.URL, time.Minute)
			action := afv(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			ai(t, res, body)
		})
	})

	t.Run("case=register and register again but login", func(t *testing.T) {
		subject = "register-twice@ory.sh"
		scope = []string{"openid"}

		t.Run("case=should pass registration", func(t *testing.T) {
			r := newRegistrationFlow(t, returnTS.URL, time.Minute)
			action := afv(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			ai(t, res, body)
		})

		t.Run("case=should pass second time registration", func(t *testing.T) {
			r := newLoginFlow(t, returnTS.URL, time.Minute)
			action := afv(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			ai(t, res, body)
		})
	})

	t.Run("case=register and complete data", func(t *testing.T) {
		subject = "incomplete-data@ory.sh"
		scope = []string{"openid"}

		t.Run("case=should fail registration on first attempt", func(t *testing.T) {
			r := newRegistrationFlow(t, returnTS.URL, time.Minute)
			action := afv(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{"traits.name": {"i"}})
			require.Contains(t, res.Request.URL.String(), uiTS.URL, "%s", body)

			assert.Equal(t, "length must be >= 2, but got 1", gjson.GetBytes(body, "methods.oidc.config.fields.#(name==traits.name).messages.0.text").String(), "%s", body) // make sure the field is being echoed
			assert.Equal(t, "traits.name", gjson.GetBytes(body, "methods.oidc.config.fields.#(name==traits.name).name").String(), "%s", body)                               // make sure the field is being echoed
			assert.Equal(t, "i", gjson.GetBytes(body, "methods.oidc.config.fields.#(name==traits.name).value").String(), "%s", body)                                        // make sure the field is being echoed
		})

		t.Run("case=should pass registration with valid data", func(t *testing.T) {
			r := newRegistrationFlow(t, returnTS.URL, time.Minute)
			action := afv(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{"traits.name": {"valid-name"}})
			ai(t, res, body)
		})
	})

	t.Run("case=should fail to register if email is already being used by password credentials", func(t *testing.T) {
		subject = "email-exist-with-password-strategy@ory.sh"
		scope = []string{"openid"}

		t.Run("case=create password identity", func(t *testing.T) {
			i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			i.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
				Identifiers: []string{subject},
			})
			i.Traits = identity.Traits(`{"subject":"` + subject + `"}`)

			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
		})

		t.Run("case=should fail registration", func(t *testing.T) {
			r := newRegistrationFlow(t, returnTS.URL, time.Minute)
			action := afv(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			aue(t, res, body, "An account with the same identifier (email, phone, username, ...) exists already.")
		})

		t.Run("case=should fail login", func(t *testing.T) {
			r := newLoginFlow(t, returnTS.URL, time.Minute)
			action := afv(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			aue(t, res, body, "An account with the same identifier (email, phone, username, ...) exists already.")
		})
	})

	t.Run("case=should redirect to default return ts when sending authenticated login flow without forced flag", func(t *testing.T) {
		subject = "no-reauth-login@ory.sh"
		scope = []string{"openid"}

		fv := url.Values{"traits.name": {"valid-name"}}
		jar, _ := cookiejar.New(nil)
		r1 := newLoginFlow(t, returnTS.URL, time.Minute)
		res1, body1 := makeRequestWithCookieJar(t, "valid", afv(t, r1.ID, "valid"), fv, jar)
		ai(t, res1, body1)
		r2 := newLoginFlow(t, returnTS.URL, time.Minute)
		res2, body2 := makeRequestWithCookieJar(t, "valid", afv(t, r2.ID, "valid"), fv, jar)
		ai(t, res2, body2)
		assert.Equal(t, body1, body2)
	})

	t.Run("case=should reauthenticate when sending authenticated login flow with forced flag", func(t *testing.T) {
		subject = "reauth-login@ory.sh"
		scope = []string{"openid"}

		fv := url.Values{"traits.name": {"valid-name"}}
		jar, _ := cookiejar.New(nil)
		r1 := newLoginFlow(t, returnTS.URL, time.Minute)
		res1, body1 := makeRequestWithCookieJar(t, "valid", afv(t, r1.ID, "valid"), fv, jar)
		ai(t, res1, body1)
		r2 := newLoginFlow(t, returnTS.URL, time.Minute)
		require.NoError(t, reg.LoginFlowPersister().ForceLoginFlow(context.Background(), r2.ID))
		res2, body2 := makeRequestWithCookieJar(t, "valid", afv(t, r2.ID, "valid"), fv, jar)
		ai(t, res2, body2)
		assert.NotEqual(t, gjson.GetBytes(body1, "id"), gjson.GetBytes(body2, "id"))
		authAt1, err := time.Parse(time.RFC3339, gjson.GetBytes(body1, "authenticated_at").String())
		require.NoError(t, err)
		authAt2, err := time.Parse(time.RFC3339, gjson.GetBytes(body2, "authenticated_at").String())
		require.NoError(t, err)
		// authenticated at is newer in the second body
		assert.Greater(t, authAt2.Sub(authAt1).Milliseconds(), int64(0), "%s - %s : %s - %s", authAt2, authAt1, body2, body1)
	})

	t.Run("method=TestPopulateSignUpMethod", func(t *testing.T) {
		conf.MustSet(config.ViperKeyPublicBaseURL, "https://foo/")

		sr := registration.NewFlow(time.Minute, "nosurf", &http.Request{URL: urlx.ParseOrPanic("/")}, flow.TypeBrowser)
		require.NoError(t, reg.RegistrationStrategies().MustStrategy(identity.CredentialsTypeOIDC).(*oidc.Strategy).PopulateRegistrationMethod(&http.Request{}, sr))

		expected := &registration.FlowMethod{
			Method: identity.CredentialsTypeOIDC,
			Config: &registration.FlowMethodConfig{
				FlowMethodConfigurator: &oidc.FlowMethod{
					HTMLForm: &form.HTMLForm{
						Action: "https://foo" + strings.ReplaceAll(oidc.RouteAuth, ":flow", sr.ID.String()),
						Method: "POST",
						Fields: form.Fields{
							{
								Name:     "csrf_token",
								Type:     "hidden",
								Required: true,
								Value:    x.FakeCSRFToken,
							},
							{
								Name:  "provider",
								Type:  "submit",
								Value: "valid",
							},
							{
								Name:  "provider",
								Type:  "submit",
								Value: "invalid-issuer",
							},
						},
					},
				},
			},
		}

		actual := sr.Methods[identity.CredentialsTypeOIDC]
		assert.EqualValues(t, expected.Config.FlowMethodConfigurator.(*oidc.FlowMethod).HTMLForm, actual.Config.FlowMethodConfigurator.(*oidc.FlowMethod).HTMLForm)
	})

	t.Run("method=TestPopulateLoginMethod", func(t *testing.T) {
		conf.MustSet(config.ViperKeyPublicBaseURL, "https://foo/")

		sr := login.NewFlow(time.Minute, "nosurf", &http.Request{URL: urlx.ParseOrPanic("/")}, flow.TypeBrowser)
		require.NoError(t, reg.LoginStrategies().MustStrategy(identity.CredentialsTypeOIDC).(*oidc.Strategy).PopulateLoginMethod(&http.Request{}, sr))

		expected := &login.FlowMethod{
			Method: identity.CredentialsTypeOIDC,
			Config: &login.FlowMethodConfig{
				FlowMethodConfigurator: &oidc.FlowMethod{
					HTMLForm: &form.HTMLForm{
						Action: "https://foo" + strings.ReplaceAll(oidc.RouteAuth, ":flow", sr.ID.String()),
						Method: "POST",
						Fields: form.Fields{
							{
								Name:     "csrf_token",
								Type:     "hidden",
								Required: true,
								Value:    x.FakeCSRFToken,
							},
							{
								Name:  "provider",
								Type:  "submit",
								Value: "valid",
							},
							{
								Name:  "provider",
								Type:  "submit",
								Value: "invalid-issuer",
							},
						},
					},
				},
			},
		}

		actual := sr.Methods[identity.CredentialsTypeOIDC]
		assert.EqualValues(t, expected.Config.FlowMethodConfigurator.(*oidc.FlowMethod).HTMLForm, actual.Config.FlowMethodConfigurator.(*oidc.FlowMethod).HTMLForm)
	})
}

func TestCountActiveCredentials(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	strategy := oidc.NewStrategy(reg)

	toJson := func(c oidc.CredentialsConfig) []byte {
		out, err := json.Marshal(&c)
		require.NoError(t, err)
		return out
	}

	for k, tc := range []struct {
		in       identity.CredentialsCollection
		expected int
	}{
		{
			in: identity.CredentialsCollection{{
				Type:   strategy.ID(),
				Config: sqlxx.JSONRawMessage{},
			}},
		},
		{
			in: identity.CredentialsCollection{{
				Type: strategy.ID(),
				Config: toJson(oidc.CredentialsConfig{Providers: []oidc.ProviderCredentialsConfig{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{""},
				Config: toJson(oidc.CredentialsConfig{Providers: []oidc.ProviderCredentialsConfig{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{"bar:"},
				Config: toJson(oidc.CredentialsConfig{Providers: []oidc.ProviderCredentialsConfig{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{":foo"},
				Config: toJson(oidc.CredentialsConfig{Providers: []oidc.ProviderCredentialsConfig{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{"not-bar:foo"},
				Config: toJson(oidc.CredentialsConfig{Providers: []oidc.ProviderCredentialsConfig{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{"bar:not-foo"},
				Config: toJson(oidc.CredentialsConfig{Providers: []oidc.ProviderCredentialsConfig{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{"bar:foo"},
				Config: toJson(oidc.CredentialsConfig{Providers: []oidc.ProviderCredentialsConfig{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
			expected: 1,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			in := make(map[identity.CredentialsType]identity.Credentials)
			for _, v := range tc.in {
				in[v.Type] = v
			}
			actual, err := strategy.CountActiveCredentials(in)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestDisabledEndpoint(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypeOIDC.String(), false)

	publicTS, _ := testhelpers.NewKratosServer(t, reg)

	t.Run("case=should not callback when oidc method is disabled", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		res, err := c.Get(publicTS.URL + oidc.RouteCallback)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		b := make([]byte, res.ContentLength)
		_, _ = res.Body.Read(b)
		assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
	})

	t.Run("case=should not auth when oidc method is disabled", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)

		t.Run("method=GET", func(t *testing.T) {
			res, err := c.Get(publicTS.URL + oidc.RouteAuth)
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)

			b := make([]byte, res.ContentLength)
			_, _ = res.Body.Read(b)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})

		t.Run("method=POST", func(t *testing.T) {
			res, err := c.PostForm(publicTS.URL+oidc.RouteAuth, url.Values{"provider": {"github"}})
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)

			b := make([]byte, res.ContentLength)
			_, _ = res.Body.Read(b)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})
	})
}
