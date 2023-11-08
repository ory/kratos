// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package profile_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ory/x/jsonx"

	kratos "github.com/ory/kratos/internal/httpclient"

	"github.com/ory/kratos/corpx"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/httpx"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/sqlxx"
)

func init() {
	corpx.RegisterFakes()
}

func newIdentityWithPassword(email string) *identity.Identity {
	return &identity.Identity{
		ID: x.NewUUID(),
		Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {Type: "password", Identifiers: []string{email}, Config: sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`)},
		},
		Traits:              identity.Traits(`{"email":"` + email + `","stringy":"foobar","booly":false,"numby":2.5,"should_long_string":"asdfasdfasdfasdfasfdasdfasdfasdf","should_big_number":2048}`),
		SchemaID:            config.DefaultIdentityTraitsSchemaID,
		State:               identity.StateActive,
		VerifiableAddresses: []identity.VerifiableAddress{{Value: email, Via: identity.VerifiableAddressTypeEmail}},
		// TO ADD - RECOVERY EMAIL,
	}
}

func TestStrategyTraits(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")
	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")
	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypePassword.String(), true)
	testhelpers.StrategyEnable(t, conf, settings.StrategyProfile, true)

	setPrivilegedTime := func(t *testing.T, duration time.Duration) {
		conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, duration.String())
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
		})
	}

	setPrivileged := func(t *testing.T) {
		setPrivilegedTime(t, time.Minute*10)
	}

	setUnprivileged := func(t *testing.T) {
		setPrivilegedTime(t, time.Nanosecond)
	}

	ui := testhelpers.NewSettingsUIEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	publicTS, _ := testhelpers.NewKratosServer(t, reg)

	browserIdentity1 := newIdentityWithPassword("john-browser@doe.com")
	apiIdentity1 := newIdentityWithPassword("john-api@doe.com")
	browserIdentity2 := &identity.Identity{ID: x.NewUUID(), Traits: identity.Traits(`{}`), State: identity.StateActive}
	apiIdentity2 := &identity.Identity{ID: x.NewUUID(), Traits: identity.Traits(`{}`), State: identity.StateActive}

	browserUser1 := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, browserIdentity1)
	browserUser2 := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, browserIdentity2)
	apiUser1 := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, apiIdentity1)
	apiUser2 := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, apiIdentity2)

	t.Run("description=not authorized to call endpoints without a session", func(t *testing.T) {
		setUnprivileged(t)

		t.Run("type=browser", func(t *testing.T) {
			res, err := http.DefaultClient.Do(httpx.MustNewRequest("POST", publicTS.URL+settings.RouteSubmitFlow, strings.NewReader(url.Values{"foo": {"bar"}}.Encode()), "application/x-www-form-urlencoded"))
			require.NoError(t, err)
			defer res.Body.Close()
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode, "%+v", res.Request)
			assert.Contains(t, res.Request.URL.String(), conf.GetProvider(ctx).String(config.ViperKeySelfServiceLoginUI))
		})

		t.Run("type=api/spa", func(t *testing.T) {
			res, err := http.DefaultClient.Do(httpx.MustNewRequest("POST", publicTS.URL+settings.RouteSubmitFlow, strings.NewReader(`{"foo":"bar"}`), "application/json"))
			require.NoError(t, err)
			defer res.Body.Close()
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
		})
	})

	t.Run("description=should fail to post data if CSRF is invalid/type=browser", func(t *testing.T) {
		setUnprivileged(t)

		f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, false, publicTS)

		actual, res := testhelpers.SettingsMakeRequest(t, false, false, f, browserUser1,
			url.Values{"traits.booly": {"true"}, "csrf_token": {"invalid"}, "method": {"profile"}}.Encode())
		assert.EqualValues(t, http.StatusOK, res.StatusCode, "should return a 400 error because CSRF token is not set\n\t%s", actual)
		assertx.EqualAsJSON(t, x.ErrInvalidCSRFToken, json.RawMessage(actual), "%s", actual)
	})

	t.Run("description=should fail to post data if CSRF is invalid/type=spa", func(t *testing.T) {
		setUnprivileged(t)

		f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, true, publicTS)

		actual, res := testhelpers.SettingsMakeRequest(t, false, true, f, browserUser1,
			testhelpers.EncodeFormAsJSON(t, true, url.Values{"traits.booly": {"true"}, "csrf_token": {"invalid"}, "method": {"profile"}}))
		assert.EqualValues(t, http.StatusForbidden, res.StatusCode, "should return a 400 error because CSRF token is not set\n\t%s", actual)
		assertx.EqualAsJSON(t, x.ErrInvalidCSRFToken, json.RawMessage(gjson.Get(actual, "error").Raw), "%s", actual)
	})

	t.Run("description=should not fail because of CSRF token but because of unprivileged/type=api", func(t *testing.T) {
		setUnprivileged(t)

		f := testhelpers.InitializeSettingsFlowViaAPI(t, apiUser1, publicTS)

		actual, res := testhelpers.SettingsMakeRequest(t, true, false, f, apiUser1, `{"traits.booly":true,"method":"profile","csrf_token":"`+x.FakeCSRFToken+`"}`)
		require.Len(t, res.Cookies(), 1)
		assert.Equal(t, "ory_kratos_continuity", res.Cookies()[0].Name)
		assert.EqualValues(t, http.StatusForbidden, res.StatusCode)
		assert.Contains(t, gjson.Get(actual, "error.reason").String(), "login session is too old", actual)
	})

	t.Run("case=should fail with correct CSRF error cause/type=api", func(t *testing.T) {
		setPrivileged(t)

		for k, tc := range []struct {
			mod func(http.Header)
			exp string
		}{
			{
				mod: func(h http.Header) {
					h.Add("Cookie", "name=bar")
				},
				exp: "The HTTP Request Header included the \\\"Cookie\\\" key",
			},
			{
				mod: func(h http.Header) {
					h.Add("Origin", "www.bar.com")
				},
				exp: "The HTTP Request Header included the \\\"Origin\\\" key",
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				f := testhelpers.InitializeSettingsFlowViaAPI(t, apiUser1, publicTS)

				req := testhelpers.NewRequest(t, true, "POST", f.Ui.Action, bytes.NewBufferString(`{"traits.booly":true,"method":"profile","csrf_token":"invalid"}`))
				tc.mod(req.Header)

				res, err := apiUser1.Do(req)
				require.NoError(t, err)
				defer res.Body.Close()

				actual := string(ioutilx.MustReadAll(res.Body))
				assert.EqualValues(t, http.StatusBadRequest, res.StatusCode)
				assert.Contains(t, actual, tc.exp, "%s", actual)
			})
		}
	})

	t.Run("description=hydrate the proper fields", func(t *testing.T) {
		setPrivileged(t)

		run := func(t *testing.T, id *identity.Identity, payload *kratos.SettingsFlow, route string) {
			assert.NotEmpty(t, payload.Identity)
			assert.Equal(t, id.ID.String(), string(payload.Identity.Id))
			assert.JSONEq(t, string(id.Traits), x.MustEncodeJSON(t, payload.Identity.Traits))
			assert.Equal(t, id.SchemaID, payload.Identity.SchemaId)
			assert.Equal(t, publicTS.URL+route, payload.RequestUrl)

			actual := jsonx.TestMarshalJSONString(t, payload.Ui)
			assert.EqualValues(t, payload.Identity.Traits.(map[string]interface{})["email"], gjson.Get(actual, "nodes.#(attributes.name==traits.email).attributes.value").String())
			assert.NotEmpty(t, gjson.Get(actual, "nodes.#(attributes.name==csrf_token).attributes.value").String(), "csrf token missing")

			testhelpers.SnapshotTExcept(t, payload.Ui, []string{"action", "nodes.0.attributes.value", "nodes.1.attributes.value"})
		}

		t.Run("type=api", func(t *testing.T) {
			pr, _, err := testhelpers.NewSDKCustomClient(publicTS, apiUser1).FrontendApi.CreateNativeSettingsFlow(context.Background()).Execute()
			require.NoError(t, err)
			run(t, apiIdentity1, pr, settings.RouteInitAPIFlow)
		})

		t.Run("type=api", func(t *testing.T) {
			pr, _, err := testhelpers.NewSDKCustomClient(publicTS, browserUser1).FrontendApi.CreateBrowserSettingsFlow(context.Background()).Execute()
			require.NoError(t, err)
			run(t, browserIdentity1, pr, settings.RouteInitBrowserFlow)
		})

		t.Run("type=browser", func(t *testing.T) {
			res, err := browserUser1.Get(publicTS.URL + settings.RouteInitBrowserFlow)
			require.NoError(t, err)
			assert.Contains(t, res.Request.URL.String(), ui.URL+"/settings?flow")

			rid := res.Request.URL.Query().Get("flow")
			require.NotEmpty(t, rid)

			pr, _, err := testhelpers.NewSDKCustomClient(publicTS, browserUser1).FrontendApi.GetSettingsFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
			require.NoError(t, err, "%s", rid)

			run(t, browserIdentity1, pr, settings.RouteInitBrowserFlow)
		})
	})

	expectValidationError := func(t *testing.T, isAPI, isSPA bool, hc *http.Client, values func(url.Values)) string {
		return testhelpers.SubmitSettingsForm(t, isAPI, isSPA, hc, publicTS, values,
			testhelpers.ExpectStatusCode(isAPI || isSPA, http.StatusBadRequest, http.StatusOK),
			testhelpers.ExpectURL(isAPI || isSPA, publicTS.URL+settings.RouteSubmitFlow, conf.SelfServiceFlowSettingsUI(ctx).String()))
	}

	t.Run("description=should come back with form errors if some profile data is invalid", func(t *testing.T) {
		setPrivileged(t)

		check := func(t *testing.T, actual string) {
			assert.NotEmpty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String(), "%s", actual)
			assert.Equal(t, "too-short", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.should_long_string).attributes.value").String(), "%s", actual)
			assert.Equal(t, "bazbar", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.stringy).attributes.value").String(), "%s", actual)
			assert.Equal(t, "2.5", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.numby).attributes.value").String(), "%s", actual)
			assert.Equal(t, "length must be >= 25, but got 9", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.should_long_string).messages.0.text").String(), "%s", actual)
		}

		payload := func(v url.Values) {
			v.Set("method", "profile")
			v.Set("traits.should_long_string", "too-short")
			v.Set("traits.stringy", "bazbar")
		}

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, apiUser1, payload))
		})

		t.Run("type=spa", func(t *testing.T) {
			check(t, expectValidationError(t, false, true, browserUser1, payload))
		})

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, false, browserUser1, payload))
		})
	})

	t.Run("description=should not be able to make requests for another user", func(t *testing.T) {
		setUnprivileged(t)

		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeSettingsFlowViaAPI(t, apiUser1, publicTS)

			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			actual, res := testhelpers.SettingsMakeRequest(t, true, false, f, apiUser2, testhelpers.EncodeFormAsJSON(t, true, values))
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "ui.messages.0.text").String(), "initiated by another person", "%s", actual)
		})

		t.Run("type=spa", func(t *testing.T) {
			f := testhelpers.InitializeSettingsFlowViaAPI(t, browserUser1, publicTS)

			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			actual, res := testhelpers.SettingsMakeRequest(t, false, true, f, browserUser2, testhelpers.EncodeFormAsJSON(t, true, values))
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "ui.messages.0.text").String(), "initiated by another person", "%s", actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, false, publicTS)

			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			actual, res := testhelpers.SettingsMakeRequest(t, false, false, f, browserUser2, values.Encode())
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "ui.messages.0.text").String(), "initiated by another person", "%s", actual)
		})
	})

	t.Run("description=should end up at the login endpoint if trying to update protected field without sudo mode", func(t *testing.T) {
		run := func(t *testing.T, config *kratos.SettingsFlow, isAPI bool, c *http.Client) *http.Response {
			time.Sleep(time.Millisecond)

			values := testhelpers.SDKFormFieldsToURLValues(config.Ui.Nodes)
			values.Set("method", "profile")
			values.Set("traits.email", "not-john-doe@foo.bar")
			res, err := c.PostForm(config.Ui.Action, values)
			require.NoError(t, err)
			defer res.Body.Close()

			return res
		}

		t.Run("type=api", func(t *testing.T) {
			setUnprivileged(t)
			f := testhelpers.InitializeSettingsFlowViaAPI(t, apiUser1, publicTS)
			res := run(t, f, true, apiUser1)
			assert.EqualValues(t, http.StatusForbidden, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
		})

		t.Run("type=sqa", func(t *testing.T) {
			setUnprivileged(t)
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, true, publicTS)
			res := run(t, f, true, browserUser1)
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), conf.GetProvider(ctx).String(config.ViperKeySelfServiceLoginUI))
		})

		t.Run("type=browser", func(t *testing.T) {
			setUnprivileged(t)
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, false, publicTS)
			res := run(t, f, false, browserUser1)
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), conf.GetProvider(ctx).String(config.ViperKeySelfServiceLoginUI))

			t.Run("should update when signed back in", func(t *testing.T) {
				setPrivileged(t)
				res, err := browserUser1.Get(f.Ui.Action)
				require.NoError(t, err)

				body := ioutilx.MustReadAll(res.Body)
				defer res.Body.Close()

				assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)
				assert.EqualValues(t, flow.StateSuccess, gjson.GetBytes(body, "state").String(), "%s", body)
			})
		})
	})

	t.Run("flow=fail first update", func(t *testing.T) {
		setPrivileged(t)

		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, flow.StateShowForm, gjson.Get(actual, "state").String(), "%s", actual)
			assert.Equal(t, "1", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.should_big_number).attributes.value").String(), "%s", actual)
			assert.Equal(t, "must be >= 1200 but found 1", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.should_big_number).messages.0.text").String(), "%s", actual)
			assert.Equal(t, "foobar", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.stringy).attributes.value").String(), "%s", actual) // sanity check if original payload is still here
		}

		payload := func(v url.Values) {
			v.Set("method", settings.StrategyProfile)
			v.Set("traits.should_big_number", "1")
		}

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, apiUser1, payload))
		})

		t.Run("type=sqa", func(t *testing.T) {
			check(t, expectValidationError(t, false, true, browserUser1, payload))
		})

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, false, browserUser1, payload))
		})
	})

	t.Run("flow=fail second update", func(t *testing.T) {
		setPrivileged(t)

		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, flow.StateShowForm, gjson.Get(actual, "state").String(), "%s", actual)

			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.should_big_number).messages.0.text").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.should_big_number).attributes.value").String(), "%s", actual)

			assert.Equal(t, "short", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.should_long_string).attributes.value").String(), "%s", actual)
			assert.Equal(t, "length must be >= 25, but got 5", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.should_long_string).messages.0.text").String(), "%s", actual)

			assert.Equal(t, "this-is-not-a-number", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.numby).attributes.value").String(), "%s", actual)
			assert.Equal(t, "expected number, but got string", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.numby).messages.0.text").String(), "%s", actual)

			assert.Equal(t, "foobar", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.stringy).attributes.value").String(), "%s", actual) // sanity check if original payload is still here
		}

		payload := func(v url.Values) {
			v.Set("method", settings.StrategyProfile)
			v.Del("traits.should_big_number")
			v.Set("traits.should_long_string", "short")
			v.Set("traits.numby", "this-is-not-a-number")
		}

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, apiUser1, payload))
		})

		t.Run("type=spa", func(t *testing.T) {
			check(t, expectValidationError(t, false, true, browserUser1, payload))
		})

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, false, browserUser1, payload))
		})
	})

	expectSuccess := func(t *testing.T, isAPI, isSPA bool, hc *http.Client, values func(url.Values)) string {
		return testhelpers.SubmitSettingsForm(t, isAPI, isSPA, hc, publicTS, values,
			http.StatusOK,
			testhelpers.ExpectURL(isAPI || isSPA, publicTS.URL+settings.RouteSubmitFlow, conf.SelfServiceFlowSettingsUI(ctx).String()))
	}

	t.Run("flow=succeed with final request", func(t *testing.T) {
		setPrivileged(t)

		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, flow.StateSuccess, gjson.Get(actual, "state").String(), "%s", actual)

			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.numby).attributes.errors").Value(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.should_big_number).attributes.errors").Value(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.should_long_string).attributes.errors").Value(), "%s", actual)

			assert.Equal(t, 15.0, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.numby).attributes.value").Value(), "%s", actual)
			assert.Equal(t, 9001.0, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.should_big_number).attributes.value").Value(), "%s", actual)
			assert.Equal(t, "this is such a long string, amazing stuff!", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.should_long_string).attributes.value").Value(), "%s", actual)
		}

		payload := func(newEmail string) func(v url.Values) {
			return func(v url.Values) {
				v.Set("method", settings.StrategyProfile)
				v.Set("traits.email", newEmail)
				v.Set("traits.numby", "15")
				v.Set("traits.should_big_number", "9001")
				v.Set("traits.should_long_string", "this is such a long string, amazing stuff!")
			}
		}

		t.Run("type=api", func(t *testing.T) {
			actual := expectSuccess(t, true, false, apiUser1, payload("not-john-doe-api@mail.com"))
			check(t, actual)
		})

		t.Run("type=sqa", func(t *testing.T) {
			actual := expectSuccess(t, false, true, browserUser1, payload("not-john-doe-browser@mail.com"))
			check(t, actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectSuccess(t, false, false, browserUser1, payload("not-john-doe-browser@mail.com")))
		})
	})

	t.Run("flow=try another update with invalid data", func(t *testing.T) {
		setPrivileged(t)

		check := func(t *testing.T, actual string) {
			assert.EqualValues(t, flow.StateShowForm, gjson.Get(actual, "state").String(), "%s", actual)
		}

		payload := func(v url.Values) {
			v.Set("method", settings.StrategyProfile)
			v.Set("traits.should_long_string", "short")
		}

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, apiUser1, payload))
		})

		t.Run("type=sqa", func(t *testing.T) {
			check(t, expectValidationError(t, false, true, browserUser1, payload))
		})

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, false, browserUser1, payload))
		})
	})

	t.Run("description=ensure that hooks are running", func(t *testing.T) {
		setPrivileged(t)

		var returned bool
		router := httprouter.New()
		router.GET("/return-ts", func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
			returned = true
		})
		rts := httptest.NewServer(router)
		t.Cleanup(rts.Close)

		testhelpers.SelfServiceHookSettingsSetDefaultRedirectTo(t, conf, rts.URL+"/return-ts")
		t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

		f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, false, publicTS)

		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("method", settings.StrategyProfile)
		values.Set("traits.should_big_number", "9001")
		res, err := browserUser1.PostForm(f.Ui.Action, values)

		require.NoError(t, err)
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.True(t, returned, "%d - %s", res.StatusCode, body)
	})

	// Update the login endpoint to auto-accept any incoming login request!
	_ = testhelpers.NewSettingsLoginAcceptAPIServer(t, testhelpers.NewSDKCustomClient(publicTS, browserUser1), conf)

	t.Run("description=should send email with verifiable address", func(t *testing.T) {
		setPrivileged(t)

		conf.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
		conf.MustSet(ctx, config.ViperKeyCourierSMTPURL, "smtp://foo:bar@irrelevant.com/")
		t.Cleanup(func() {
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceSettingsAfter, settings.StrategyProfile), nil)
		})

		check := func(t *testing.T, actual, newEmail string) {
			assert.EqualValues(t, flow.StateSuccess, gjson.Get(actual, "state").String(), "%s", actual)
			assert.Equal(t, newEmail, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.email).attributes.value").Value(), "%s", actual)

			m, err := reg.CourierPersister().LatestQueuedMessage(context.Background())
			require.NoError(t, err)
			assert.Contains(t, m.Subject, "verify your email address")
		}

		payload := func(newEmail string) func(v url.Values) {
			return func(v url.Values) {
				v.Set("method", settings.StrategyProfile)
				v.Set("traits.email", newEmail)
			}
		}

		t.Run("type=api", func(t *testing.T) {
			newEmail := "update-verify-api@mail.com"
			actual := expectSuccess(t, true, false, apiUser1, payload(newEmail))
			check(t, actual, newEmail)
		})

		t.Run("type=spa", func(t *testing.T) {
			newEmail := "update-verify-browser@mail.com"
			actual := expectSuccess(t, false, true, browserUser1, payload(newEmail))
			check(t, actual, newEmail)
		})

		t.Run("type=browser", func(t *testing.T) {
			newEmail := "update-verify-browser@mail.com"
			actual := expectSuccess(t, false, false, browserUser1, payload(newEmail))
			check(t, actual, newEmail)
		})
	})

	t.Run("description=should update protected field with sudo mode", func(t *testing.T) {
		setPrivileged(t)

		check := func(t *testing.T, newEmail string, actual string) {
			assert.EqualValues(t, flow.StateSuccess, gjson.Get(actual, "state").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.numby).attributes.errors").Value(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.should_big_number).attributes.errors").Value(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.should_long_string).attributes.errors").Value(), "%s", actual)
			assert.Equal(t, newEmail, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.email).attributes.value").Value(), "%s", actual)
			assert.Equal(t, "foobar", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.stringy).attributes.value").String(), "%s", actual) // sanity check if original payload is still here
		}

		payload := func(email string) func(v url.Values) {
			return func(v url.Values) {
				v.Set("method", settings.StrategyProfile)
				v.Set("traits.email", email)
			}
		}

		t.Run("type=api", func(t *testing.T) {
			setPrivilegedTime(t, time.Second*10)
			email := "not-john-doe-api@mail.com"
			actual := expectSuccess(t, true, false, apiUser1, payload(email))
			check(t, email, actual)
		})

		t.Run("type=sqa", func(t *testing.T) {
			setPrivilegedTime(t, time.Second*10)
			email := "not-john-doe-browser@mail.com"
			actual := expectSuccess(t, false, true, browserUser1, payload(email))
			check(t, email, actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			setPrivilegedTime(t, time.Second*10)
			email := "not-john-doe-browser@mail.com"
			actual := expectSuccess(t, false, false, browserUser1, payload(email))
			check(t, email, actual)
		})
	})
}

func TestDisabledEndpoint(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")
	testhelpers.StrategyEnable(t, conf, settings.StrategyProfile, false)
	errTS := testhelpers.NewErrorTestServer(t, reg)

	publicTS, _ := testhelpers.NewKratosServer(t, reg)
	browserIdentity1 := newIdentityWithPassword("john-browser@doe.com")
	browserUser1 := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, browserIdentity1)

	t.Run("case=should not submit when profile method is disabled", func(t *testing.T) {
		t.Run("method=GET", func(t *testing.T) {
			t.Skip("GET is not supported here")
		})

		t.Run("method=POST", func(t *testing.T) {
			b := testhelpers.SubmitSettingsForm(t, false, false, browserUser1, publicTS, func(v url.Values) {
				v.Set("method", settings.StrategyProfile)
			},
				http.StatusOK,
				testhelpers.ExpectURL(false, publicTS.URL+settings.RouteSubmitFlow, errTS.URL))

			assert.Contains(t, string(b), "404")
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})
	})
}
