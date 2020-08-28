package profile_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"

	"github.com/ory/x/sqlxx"

	"github.com/ory/x/pointerx"

	"github.com/ory/x/httpx"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/httpclient/client/common"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/selfservice/strategy/profile"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

func newIdentityWithPassword(email string) *identity.Identity {
	return &identity.Identity{
		ID: x.NewUUID(),
		Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {Type: "password", Identifiers: []string{email}, Config: sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`)},
		},
		Traits:   identity.Traits(`{"email":"` + email + `","stringy":"foobar","booly":false,"numby":2.5,"should_long_string":"asdfasdfasdfasdfasfdasdfasdfasdf","should_big_number":2048}`),
		SchemaID: configuration.DefaultIdentityTraitsSchemaID,
		VerifiableAddresses: []identity.VerifiableAddress{
			{Value: email, Via: identity.VerifiableAddressTypeEmail, Code: x.NewUUID().String()},
		},
		// TO ADD - RECOVERY EMAIL,
	}
}

func TestStrategyTraits(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")
	viper.Set(configuration.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")
	viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
	testhelpers.StrategyEnable(identity.CredentialsTypePassword.String(), true)
	testhelpers.StrategyEnable(settings.StrategyProfile, true)

	ui := testhelpers.NewSettingsUIEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	publicTS, adminTS := testhelpers.NewKratosServer(t, reg)

	browserIdentity1 := newIdentityWithPassword("john-browser@doe.com")
	apiIdentity1 := newIdentityWithPassword("john-api@doe.com")
	browserIdentity2 := &identity.Identity{ID: x.NewUUID(), Traits: identity.Traits(`{}`)}
	apiIdentity2 := &identity.Identity{ID: x.NewUUID(), Traits: identity.Traits(`{}`)}

	browserUser1 := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, browserIdentity1)
	browserUser2 := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, browserIdentity2)
	apiUser1 := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, apiIdentity1)
	apiUser2 := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, apiIdentity2)

	publicClient := testhelpers.NewSDKClient(publicTS)
	adminClient := testhelpers.NewSDKClient(adminTS)

	t.Run("description=not authorized to call endpoints without a session", func(t *testing.T) {
		t.Run("type=browser", func(t *testing.T) {
			res, err := http.DefaultClient.Do(httpx.MustNewRequest("POST", publicTS.URL+profile.RouteSettings, strings.NewReader(url.Values{"foo": {"bar"}}.Encode()), "application/x-www-form-urlencoded"))
			require.NoError(t, err)
			defer res.Body.Close()
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode, "%+v", res.Request)
			assert.Contains(t, res.Request.URL.String(), viper.GetString(configuration.ViperKeySelfServiceLoginUI))
		})

		t.Run("type=api", func(t *testing.T) {
			res, err := http.DefaultClient.Do(httpx.MustNewRequest("POST", publicTS.URL+profile.RouteSettings, strings.NewReader(`{"foo":"bar"}`), "application/json"))
			require.NoError(t, err)
			defer res.Body.Close()
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
		})
	})

	t.Run("description=should fail to post data if CSRF is invalid/type=browser", func(t *testing.T) {
		rs := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, publicTS)
		f := testhelpers.GetSettingsFlowMethodConfig(t, rs.Payload, settings.StrategyProfile)

		actual, res := testhelpers.SettingsMakeRequest(t, false, f, browserUser1,
			url.Values{"traits.foo": {"bar"}, "csrf_token": {"invalid"}}.Encode())
		assert.EqualValues(t, http.StatusOK, res.StatusCode, "should return a 400 error because CSRF token is not set\n\t%s", actual)
		assertx.EqualAsJSON(t, x.ErrInvalidCSRFToken, json.RawMessage(gjson.Get(actual, "0").Raw))
	})

	t.Run("description=should not fail if CSRF token is invalid/type=api", func(t *testing.T) {
		rs := testhelpers.InitializeSettingsFlowViaAPI(t, browserUser1, publicTS)
		f := testhelpers.GetSettingsFlowMethodConfig(t, rs.Payload, settings.StrategyProfile)

		actual, res := testhelpers.SettingsMakeRequest(t, true, f, browserUser1, `{"traits":{},"csrf_token":"invalid"}`)
		assert.EqualValues(t, http.StatusBadRequest, res.StatusCode)
		assert.EqualValues(t, "api", gjson.Get(actual, "type").String())
	})

	t.Run("description=hydrate the proper fields", func(t *testing.T) {
		var run = func(t *testing.T, id *identity.Identity, payload *models.SettingsFlow, route string) {
			assert.NotEmpty(t, payload.Identity)
			assert.Equal(t, id.ID.String(), string(payload.Identity.ID))
			assert.JSONEq(t, string(id.Traits), x.MustEncodeJSON(t, payload.Identity.Traits))
			assert.Equal(t, id.SchemaID, pointerx.StringR(payload.Identity.SchemaID))
			assert.Equal(t, publicTS.URL+route, pointerx.StringR(payload.RequestURL))

			f := testhelpers.GetSettingsFlowMethodConfig(t, payload, settings.StrategyProfile)

			assertx.EqualAsJSON(t, &models.Form{
				Action: pointerx.String(publicTS.URL + profile.RouteSettings + "?flow=" + string(payload.ID)),
				Method: pointerx.String("POST"),
				Fields: models.FormFields{
					&models.FormField{Name: pointerx.String(form.CSRFTokenName), Required: true, Type: pointerx.String("hidden"), Value: x.FakeCSRFToken},
					&models.FormField{Name: pointerx.String("traits.email"), Type: pointerx.String("text"), Value: gjson.GetBytes(id.Traits, "email").String()},
					&models.FormField{Name: pointerx.String("traits.stringy"), Type: pointerx.String("text"), Value: "foobar"},
					&models.FormField{Name: pointerx.String("traits.numby"), Type: pointerx.String("number"), Value: json.Number("2.5")},
					&models.FormField{Name: pointerx.String("traits.booly"), Type: pointerx.String("checkbox"), Value: false},
					&models.FormField{Name: pointerx.String("traits.should_big_number"), Type: pointerx.String("number"), Value: json.Number("2048")},
					&models.FormField{Name: pointerx.String("traits.should_long_string"), Type: pointerx.String("text"), Value: "asdfasdfasdfasdfasfdasdfasdfasdf"},
				},
			}, f)
		}

		t.Run("type=api", func(t *testing.T) {
			pr, err := publicClient.Common.InitializeSelfServiceSettingsViaAPIFlow(
				common.NewInitializeSelfServiceSettingsViaAPIFlowParams().WithHTTPClient(apiUser1))
			require.NoError(t, err)
			run(t, apiIdentity1, pr.Payload, settings.RouteInitAPIFlow)
		})

		t.Run("type=browser", func(t *testing.T) {
			res, err := browserUser1.Get(publicTS.URL + settings.RouteInitBrowserFlow)
			require.NoError(t, err)
			assert.Contains(t, res.Request.URL.String(), ui.URL+"/settings?flow")

			rid := res.Request.URL.Query().Get("flow")
			require.NotEmpty(t, rid)

			pr, err := publicClient.Common.GetSelfServiceSettingsFlow(
				common.NewGetSelfServiceSettingsFlowParams().WithHTTPClient(browserUser1).
					WithID(res.Request.URL.Query().Get("flow")))
			require.NoError(t, err, "%s", rid)

			run(t, browserIdentity1, pr.Payload, settings.RouteInitBrowserFlow)
		})
	})

	var expectValidationError = func(t *testing.T, isAPI bool, hc *http.Client, values func(url.Values)) string {
		return testhelpers.SubmitSettingsForm(t, isAPI, hc, publicTS, values,
			settings.StrategyProfile,
			testhelpers.ExpectStatusCode(isAPI, http.StatusBadRequest, http.StatusOK),
			testhelpers.ExpectURL(isAPI, publicTS.URL+profile.RouteSettings, conf.SelfServiceFlowSettingsUI().String()))
	}

	t.Run("description=should come back with form errors if some profile data is invalid", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.NotEmpty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==csrf_token).value").String(), "%s", actual)
			assert.Equal(t, "too-short", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_long_string).value").String(), "%s", actual)
			assert.Equal(t, "bazbar", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.stringy).value").String(), "%s", actual)
			assert.Equal(t, "2.5", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.numby).value").String(), "%s", actual)
			assert.Equal(t, "length must be >= 25, but got 9", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_long_string).messages.0.text").String(), "%s", actual)
		}

		var payload = func(v url.Values) {
			v.Set("traits.should_long_string", "too-short")
			v.Set("traits.stringy", "bazbar")
		}

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, apiUser1, payload))
		})

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, browserUser1, payload))
		})
	})

	t.Run("description=should not be able to make requests for another user", func(t *testing.T) {
		t.Run("type=api", func(t *testing.T) {
			rs := testhelpers.InitializeSettingsFlowViaAPI(t, apiUser1, publicTS)
			f := testhelpers.GetSettingsFlowMethodConfig(t, rs.Payload, identity.CredentialsTypePassword.String())
			values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
			actual, res := testhelpers.SettingsMakeRequest(t, true, f, apiUser2, testhelpers.EncodeFormAsJSON(t, true, values))
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "error.reason").String(), "initiated by another person", "%s", actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			rs := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, publicTS)
			f := testhelpers.GetSettingsFlowMethodConfig(t, rs.Payload, identity.CredentialsTypePassword.String())
			values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
			actual, res := testhelpers.SettingsMakeRequest(t, false, f, browserUser2, values.Encode())
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "0.reason").String(), "initiated by another person", "%s", actual)
		})
	})

	t.Run("description=should end up at the login endpoint if trying to update protected field without sudo mode", func(t *testing.T) {
		var run = func(t *testing.T, config *models.FlowMethodConfig, isAPI bool, c *http.Client) *http.Response {
			time.Sleep(time.Millisecond)

			values := testhelpers.SDKFormFieldsToURLValues(config.Fields)
			values.Set("traits.email", "not-john-doe@foo.bar")
			res, err := c.PostForm(pointerx.StringR(config.Action), values)
			require.NoError(t, err)
			defer res.Body.Close()

			return res
		}

		t.Run("type=api", func(t *testing.T) {
			rs := testhelpers.InitializeSettingsFlowViaAPI(t, apiUser1, publicTS)
			config := testhelpers.GetSettingsFlowMethodConfig(t, rs.Payload, settings.StrategyProfile)
			res := run(t, config, true, apiUser1)
			assert.EqualValues(t, http.StatusForbidden, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+profile.RouteSettings)
		})

		t.Run("type=browser", func(t *testing.T) {
			rs := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, publicTS)
			config := testhelpers.GetSettingsFlowMethodConfig(t, rs.Payload, settings.StrategyProfile)
			res := run(t, config, false, browserUser1)
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), viper.Get(configuration.ViperKeySelfServiceLoginUI))
		})
	})

	t.Run("flow=fail first update", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.EqualValues(t, settings.StateShowForm, gjson.Get(actual, "state").String(), "%s", actual)
			assert.Equal(t, "1", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_big_number).value").String(), "%s", actual)
			assert.Equal(t, "must be >= 1200 but found 1", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_big_number).messages.0.text").String(), "%s", actual)
			assert.Equal(t, "foobar", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.stringy).value").String(), "%s", actual) // sanity check if original payload is still here
		}

		var payload = func(v url.Values) {
			v.Set("traits.should_big_number", "1")
		}

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, apiUser1, payload))
		})

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, browserUser1, payload))
		})
	})

	t.Run("flow=fail second update", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.EqualValues(t, settings.StateShowForm, gjson.Get(actual, "state").String(), "%s", actual)

			assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_big_number).messages.0.text").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_big_number).value").String(), "%s", actual)

			assert.Equal(t, "short", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_long_string).value").String(), "%s", actual)
			assert.Equal(t, "length must be >= 25, but got 5", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_long_string).messages.0.text").String(), "%s", actual)

			assert.Equal(t, "this-is-not-a-number", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.numby).value").String(), "%s", actual)
			assert.Equal(t, "expected number, but got string", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.numby).messages.0.text").String(), "%s", actual)

			assert.Equal(t, "foobar", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.stringy).value").String(), "%s", actual) // sanity check if original payload is still here
		}

		var payload = func(v url.Values) {
			v.Del("traits.should_big_number")
			v.Set("traits.should_long_string", "short")
			v.Set("traits.numby", "this-is-not-a-number")
		}

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, apiUser1, payload))
		})

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, browserUser1, payload))
		})
	})

	var expectSuccess = func(t *testing.T, isAPI bool, hc *http.Client, values func(url.Values)) string {
		return testhelpers.SubmitSettingsForm(t, isAPI, hc, publicTS, values,
			settings.StrategyProfile, http.StatusOK,
			testhelpers.ExpectURL(isAPI, publicTS.URL+profile.RouteSettings, conf.SelfServiceFlowSettingsUI().String()))
	}

	t.Run("flow=succeed with final request", func(t *testing.T) {
		viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1h")
		defer viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")

		var check = func(t *testing.T, actual string) {
			assert.EqualValues(t, settings.StateSuccess, gjson.Get(actual, "state").String(), "%s", actual)

			assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.numby).errors").Value(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_big_number).errors").Value(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_long_string).errors").Value(), "%s", actual)

			assert.Equal(t, 15.0, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.numby).value").Value(), "%s", actual)
			assert.Equal(t, 9001.0, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_big_number).value").Value(), "%s", actual)
			assert.Equal(t, "this is such a long string, amazing stuff!", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_long_string).value").Value(), "%s", actual)
		}

		var payload = func(newEmail string) func(v url.Values) {
			return func(v url.Values) {
				v.Set("traits.email", newEmail)
				v.Set("traits.numby", "15")
				v.Set("traits.should_big_number", "9001")
				v.Set("traits.should_long_string", "this is such a long string, amazing stuff!")
			}
		}

		t.Run("type=api", func(t *testing.T) {
			actual := expectSuccess(t, true, apiUser1, payload("not-john-doe-api@mail.com"))
			check(t, gjson.Get(actual, "flow").Raw)
		})

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectSuccess(t, false, browserUser1, payload("not-john-doe-browser@mail.com")))
		})
	})

	t.Run("flow=try another update with invalid data", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.EqualValues(t, settings.StateShowForm, gjson.Get(actual, "state").String(), "%s", actual)
		}

		var payload = func(v url.Values) {
			v.Set("traits.should_long_string", "short")
		}

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, apiUser1, payload))
		})

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, browserUser1, payload))
		})
	})

	t.Run("description=ensure that hooks are running", func(t *testing.T) {
		var returned bool
		router := httprouter.New()
		router.GET("/return-ts", func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
			returned = true
		})
		rts := httptest.NewServer(router)
		t.Cleanup(rts.Close)

		testhelpers.SelfServiceHookSettingsSetDefaultRedirectTo(rts.URL + "/return-ts")
		t.Cleanup(testhelpers.SelfServiceHookConfigReset)

		f := testhelpers.GetSettingsFlowMethodConfigDeprecated(t, browserUser1, publicTS, settings.StrategyProfile)

		values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
		values.Set("traits.should_big_number", "9001")
		res, err := browserUser1.PostForm(pointerx.StringR(f.Action), values)

		require.NoError(t, err)
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		assert.True(t, returned, "%d - %s", res.StatusCode, body)
	})

	// Update the login endpoint to auto-accept any incoming login request!
	_ = testhelpers.NewSettingsLoginAcceptAPIServer(t, adminClient)

	t.Run("description=should send email with verifiable address", func(t *testing.T) {
		viper.Set(configuration.ViperKeySelfServiceVerificationEnabled, true)
		viper.Set(configuration.ViperKeyCourierSMTPURL, "smtp://foo:bar@irrelevant.com/")
		viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1h")
		defer viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
		defer viper.Set(configuration.HookStrategyKey(configuration.ViperKeySelfServiceSettingsAfter, settings.StrategyProfile), nil)

		var check = func(t *testing.T, actual, newEmail string) {
			assert.EqualValues(t, settings.StateSuccess, gjson.Get(actual, "state").String(), "%s", actual)
			assert.Equal(t, newEmail, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.email).value").Value(), "%s", actual)

			m, err := reg.CourierPersister().LatestQueuedMessage(context.Background())
			require.NoError(t, err)
			assert.Contains(t, m.Subject, "verify your email address")
		}

		var payload = func(newEmail string) func(v url.Values) {
			return func(v url.Values) {
				v.Set("traits.email", newEmail)
			}
		}

		t.Run("type=api", func(t *testing.T) {
			newEmail := "update-verify-api@mail.com"
			actual := expectSuccess(t, true, apiUser1, payload(newEmail))
			check(t, gjson.Get(actual, "flow").String(), newEmail)
		})

		t.Run("type=browser", func(t *testing.T) {
			newEmail := "update-verify-browser@mail.com"
			actual := expectSuccess(t, false, browserUser1, payload(newEmail))
			check(t, actual, newEmail)
		})
	})

	t.Run("description=should update protected field with sudo mode", func(t *testing.T) {
		viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")
		defer viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")

		var check = func(t *testing.T, newEmail string, actual string) {
			assert.EqualValues(t, settings.StateSuccess, gjson.Get(actual, "state").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.numby).errors").Value(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_big_number).errors").Value(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.should_long_string).errors").Value(), "%s", actual)
			assert.Equal(t, newEmail, gjson.Get(actual, "methods.profile.config.fields.#(name==traits.email).value").Value(), "%s", actual)
			assert.Equal(t, "foobar", gjson.Get(actual, "methods.profile.config.fields.#(name==traits.stringy).value").String(), "%s", actual) // sanity check if original payload is still here
		}

		var payload = func(email string) func(v url.Values) {
			return func(v url.Values) {
				v.Set("traits.email", email)
			}
		}

		t.Run("type=api", func(t *testing.T) {
			email := "not-john-doe-api@mail.com"
			actual := expectSuccess(t, true, apiUser1, payload(email))
			check(t, email, gjson.Get(actual, "flow").Raw)
		})

		t.Run("type=browser", func(t *testing.T) {
			email := "not-john-doe-browser@mail.com"
			actual := expectSuccess(t, false, browserUser1, payload(email))
			check(t, email, actual)
		})
	})
}
