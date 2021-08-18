package lookup_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/lookup"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
)

//go:embed fixtures/login/with.json
var loginFixtureWithLookup []byte

var lookupCodeGJSONQuery = "ui.nodes.#(attributes.name==" + identity.CredentialsTypeLookup.String() + ")"

func TestCompleteLogin(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword)+".enabled", false)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeLookup)+".enabled", true)

	router := x.NewRouterPublic()
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	errTS := testhelpers.NewErrorTestServer(t, reg)
	uiTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)
	redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

	// Overwrite these two to make it more explicit when tests fail
	conf.MustSet(config.ViperKeySelfServiceErrorUI, errTS.URL+"/error-ts")
	conf.MustSet(config.ViperKeySelfServiceLoginUI, uiTS.URL+"/login-ts")

	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/login.schema.json")
	conf.MustSet(config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

	t.Run("case=lookup payload is set when identity has lookup", func(t *testing.T) {
		id, _ := createIdentity(t, reg)

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
		assertx.EqualAsJSONExcept(t, json.RawMessage(loginFixtureWithLookup), f.Ui.Nodes, []string{"2.attributes.value"})
	})

	t.Run("case=lookup payload is not set when identity has no lookup", func(t *testing.T) {
		id := createIdentityWithoutLookup(t, reg)

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
		assertx.EqualAsJSON(t, nil, f.Ui.Nodes)
	})

	t.Run("case=lookup payload is not set when identity has no lookup", func(t *testing.T) {
		id := createIdentityWithoutLookup(t, reg)

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
		assertx.EqualAsJSON(t, nil, f.Ui.Nodes)
	})

	doAPIFlowWithClient := func(t *testing.T, v func(url.Values), id *identity.Identity, apiClient *http.Client, forced bool) (string, *http.Response) {
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, forced, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("method", identity.CredentialsTypeLookup.String())
		v(values)
		payload := testhelpers.EncodeFormAsJSON(t, true, values)
		return testhelpers.LoginMakeRequest(t, true, false, f, apiClient, payload)
	}

	doAPIFlow := func(t *testing.T, v func(url.Values), id *identity.Identity) (string, *http.Response) {
		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		return doAPIFlowWithClient(t, v, id, apiClient, false)
	}

	doBrowserFlowWithClient := func(t *testing.T, spa bool, v func(url.Values), id *identity.Identity, browserClient *http.Client, forced bool) (string, *http.Response) {
		f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, forced, spa, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("method", identity.CredentialsTypeLookup.String())
		v(values)
		return testhelpers.LoginMakeRequest(t, false, spa, f, browserClient, values.Encode())
	}

	doBrowserFlow := func(t *testing.T, spa bool, v func(url.Values), id *identity.Identity) (string, *http.Response) {
		browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
		return doBrowserFlowWithClient(t, spa, v, id, browserClient, false)
	}

	checkURL := func(t *testing.T, shouldRedirect bool, res *http.Response) {
		if shouldRedirect {
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login-ts")
		} else {
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
		}
	}

	t.Run("case=should fail if code is invalid", func(t *testing.T) {
		for _, tc := range []struct {
			d       string
			code    string
			message string
		}{
			{d: "empty", message: text.NewErrorValidationLookupInvalid().Text},
			{d: "invalid", code: "invalid", message: text.NewErrorValidationLookupInvalid().Text},
			{d: "already-used", code: "key-1", message: text.NewErrorValidationLookupAlreadyUsed().Text},
		} {
			t.Run(fmt.Sprintf("code=%s", tc.d), func(t *testing.T) {
				id, _ := createIdentity(t, reg)
				payload := func(v url.Values) {
					v.Set(node.LookupCodeEnter, tc.code)
				}

				check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response) {
					checkURL(t, shouldRedirect, res)
					assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
					assert.Equal(t, tc.message, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
				}

				t.Run("type=api", func(t *testing.T) {
					body, res := doAPIFlow(t, payload, id)
					check(t, false, body, res)
				})

				t.Run("type=browser", func(t *testing.T) {
					body, res := doBrowserFlow(t, false, payload, id)
					check(t, true, body, res)
				})

				t.Run("type=spa", func(t *testing.T) {
					body, res := doBrowserFlow(t, true, payload, id)
					check(t, false, body, res)
				})
			})
		}
	})

	t.Run("case=should fail if lookup was not set up for identity", func(t *testing.T) {
		id := createIdentityWithoutLookup(t, reg)

		payload := func(v url.Values) {
			v.Set(node.LookupCodeEnter, "1111111")
		}

		check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response) {
			checkURL(t, shouldRedirect, res)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Equal(t, text.NewErrorValidationNoLookup().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
		}

		t.Run("type=api", func(t *testing.T) {
			body, res := doAPIFlow(t, payload, id)
			check(t, false, body, res)
		})

		t.Run("type=browser", func(t *testing.T) {
			body, res := doBrowserFlow(t, false, payload, id)
			check(t, true, body, res)
		})

		t.Run("type=spa", func(t *testing.T) {
			body, res := doBrowserFlow(t, true, payload, id)
			check(t, false, body, res)
		})
	})

	t.Run("case=should pass when code is supplied correctly", func(t *testing.T) {
		id, _ := createIdentity(t, reg)
		payload := func(code string) func(v url.Values) {
			return func(v url.Values) {
				v.Set(node.LookupCodeEnter, code)
			}
		}

		startAt := time.Now()
		check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response, usedKey string, expectedAuths int) {
			prefix := "session."
			if shouldRedirect {
				assert.Contains(t, res.Request.URL.String(), redirTS.URL+"/return-ts")
				prefix = ""
			} else {
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			}
			assert.True(t, gjson.Get(body, prefix+"active").Bool(), "%s", body)
			assert.EqualValues(t, identity.AuthenticatorAssuranceLevel2, gjson.Get(body, prefix+"authenticator_assurance_level").String())
			require.Len(t, gjson.Get(body, prefix+"authentication_methods").Array(), expectedAuths)
			assert.EqualValues(t, identity.CredentialsTypePassword, gjson.Get(body, prefix+"authentication_methods.0.method").String())
			assert.True(t, gjson.Get(body, prefix+"authentication_methods.0.completed_at").Time().After(startAt))
			assert.EqualValues(t, identity.CredentialsTypeLookup, gjson.Get(body, prefix+"authentication_methods.1.method").String())
			assert.True(t, gjson.Get(body, prefix+"authentication_methods.1.completed_at").Time().After(startAt))
			assert.True(t, gjson.Get(body, prefix+"authentication_methods.1.completed_at").Time().After(gjson.Get(body, prefix+"authentication_methods.0.completed_at").Time()))
			if expectedAuths == 3 {
				assert.EqualValues(t, identity.CredentialsTypeLookup, gjson.Get(body, prefix+"authentication_methods.2.method").String())
				assert.True(t, gjson.Get(body, prefix+"authentication_methods.2.completed_at").Time().After(startAt))
				assert.True(t, gjson.Get(body, prefix+"authentication_methods.2.completed_at").Time().After(gjson.Get(body, prefix+"authentication_methods.1.completed_at").Time()))
				assert.False(t, gjson.Get(body, prefix+"authenticated_at").Time().Before(gjson.Get(body, prefix+"authentication_methods.2.completed_at").Time()))
			} else {
				assert.False(t, gjson.Get(body, prefix+"authenticated_at").Time().Before(gjson.Get(body, prefix+"authentication_methods.1.completed_at").Time()))
			}

			actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), uuid.FromStringOrNil(gjson.Get(body, prefix+"identity.id").String()))
			require.NoError(t, err)
			creds, ok := actual.GetCredentials(identity.CredentialsTypeLookup)
			require.True(t, ok)

			var conf lookup.CredentialsConfig
			require.NoError(t, json.Unmarshal(creds.Config, &conf))

			var found bool
			for _, rc := range conf.RecoveryCodes {
				if rc.Code == usedKey {
					found = true
					require.False(t, time.Time(rc.UsedAt).IsZero())
				}
			}

			require.True(t, found)
		}

		t.Run("type=api", func(t *testing.T) {
			apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
			body, res := doAPIFlowWithClient(t, payload("key-0"), id, apiClient, false)
			check(t, false, body, res, "key-0", 2)
			// We can still use another key
			body, res = doAPIFlowWithClient(t, payload("key-2"), id, apiClient, true)
			check(t, false, body, res, "key-2", 3)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
			body, res := doBrowserFlowWithClient(t, false, payload("key-3"), id, browserClient, false)
			check(t, true, body, res, "key-3", 2)
			// We can still use another key
			body, res = doBrowserFlowWithClient(t, false, payload("key-5"), id, browserClient, true)
			check(t, true, body, res, "key-5", 3)
		})

		t.Run("type=spa", func(t *testing.T) {
			browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
			body, res := doBrowserFlowWithClient(t, true, payload("key-6"), id, browserClient, false)
			check(t, false, body, res, "key-6", 2)
			// We can still use another key
			body, res = doBrowserFlowWithClient(t, true, payload("key-8"), id, browserClient, true)
			check(t, false, body, res, "key-8", 3)
		})
	})

	t.Run("case=should fail because lookup can not handle AAL1", func(t *testing.T) {
		apiClient := testhelpers.NewDebugClient(t)
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

		update, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(f.Id))
		require.NoError(t, err)
		update.RequestedAAL = identity.AuthenticatorAssuranceLevel1
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), update))

		req, err := http.NewRequest("POST", f.Ui.Action, bytes.NewBufferString(`{"method":"lookup"}`))
		require.NoError(t, err)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		body := x.MustReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
		assert.Equal(t, text.NewErrorValidationLoginNoStrategyFound().Text, gjson.GetBytes(body, "ui.messages.0.text").String())
	})

	t.Run("case=should pass without csrf if API flow", func(t *testing.T) {
		id, _ := createIdentity(t, reg)
		body, res := doAPIFlow(t, func(v url.Values) {
			v.Del("csrf_token")
			v.Set(node.LookupCodeEnter, "111111")
		}, id)

		assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
		assert.Equal(t, text.NewErrorValidationLookupInvalid().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
	})

	t.Run("case=should fail if CSRF token is invalid", func(t *testing.T) {
		id, _ := createIdentity(t, reg)
		t.Run("type=browser", func(t *testing.T) {
			body, res := doBrowserFlow(t, false, func(v url.Values) {
				v.Del("csrf_token")
				v.Set(node.LookupCodeEnter, "111111")
			}, id)

			assert.Contains(t, res.Request.URL.String(), errTS.URL)
			assert.Equal(t, x.ErrInvalidCSRFToken.Reason(), gjson.Get(body, "reason").String(), body)
		})

		t.Run("type=spa", func(t *testing.T) {
			body, res := doBrowserFlow(t, true, func(v url.Values) {
				v.Del("csrf_token")
				v.Set(node.LookupCodeEnter, "111111")
			}, id)

			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.Equal(t, x.ErrInvalidCSRFToken.Reason(), gjson.Get(body, "error.reason").String(), body)
		})
	})
}
