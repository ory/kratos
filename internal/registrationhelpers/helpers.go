package registrationhelpers

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	kratos "github.com/ory/kratos-client-go"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/httpx"
	"github.com/ory/x/ioutilx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func setupServer(t *testing.T, reg *driver.RegistryDefault) *httptest.Server {
	conf := reg.Config(context.Background())
	router := x.NewRouterPublic()
	admin := x.NewRouterAdmin()

	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, admin)
	redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)
	conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, redirTS.URL+"/default-return-to")
	conf.MustSet(config.ViperKeySelfServiceRegistrationAfter+"."+config.DefaultBrowserReturnURL, redirTS.URL+"/registration-return-ts")
	return publicTS
}

func ExpectValidationError(t *testing.T, ts *httptest.Server, conf *config.Config, flow string, values func(url.Values)) string {
	isSPA := flow == "spa"
	isAPI := flow == "api"
	return testhelpers.SubmitRegistrationForm(t, isAPI, nil, ts, values,
		isSPA,
		testhelpers.ExpectStatusCode(isAPI || isSPA, http.StatusBadRequest, http.StatusOK),
		testhelpers.ExpectURL(isAPI || isSPA, ts.URL+registration.RouteSubmitFlow, conf.SelfServiceFlowRegistrationUI().String()))
}

func CheckFormContent(t *testing.T, body []byte, requiredFields ...string) {
	FieldNameSet(t, body, requiredFields)
	OutdatedFieldsDoNotExist(t, body)
	FormMethodIsPOST(t, body)
}

// FieldNameSet checks if the fields have the right "name" set.
func FieldNameSet(t *testing.T, body []byte, fields []string) {
	for _, f := range fields {
		assert.Equal(t, f, gjson.GetBytes(body, fmt.Sprintf("ui.nodes.#(attributes.name==%s).attributes.name", f)).String(), "%s", body)
	}
}

// checks if some keys are not set, this should be used to catch regression issues
func OutdatedFieldsDoNotExist(t *testing.T, body []byte) {
	for _, k := range []string{"request"} {
		assert.Equal(t, false, gjson.GetBytes(body, fmt.Sprintf("ui.nodes.fields.#(name==%s)", k)).Exists())
	}
}

func FormMethodIsPOST(t *testing.T, body []byte) {
	assert.Equal(t, "POST", gjson.GetBytes(body, "ui.method").String())
}

//go:embed stub/basic.schema.json
var basicSchema []byte

//go:embed stub/multifield.schema.json
var multifieldSchema []byte

func AssertRegistrationRespectsValidation(t *testing.T, flows []string, payload func(url.Values)) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchemaFromRaw(conf, multifieldSchema)
	_ = testhelpers.NewRegistrationUIFlowEchoServer(t, reg)
	publicTS := setupServer(t, reg)

	t.Run("case=should return an error because not passing validation", func(t *testing.T) {
		email := testhelpers.RandomEmail()
		var check = func(t *testing.T, actual string) {
			assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "ui.action").String(), publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
			CheckFormContent(t, []byte(actual), "password", "csrf_token", "traits.username", "traits.foobar")
			assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.foobar).messages.0").String(), `Property foobar is missing`, "%s", actual)
			assert.Equal(t, email, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).attributes.value").String(), "%s", actual)
		}

		var values = func(v url.Values) {
			v.Set("traits.username", email)
			v.Del("traits.foobar")
			payload(v)
		}

		for _, f := range flows {
			t.Run("type="+f, func(t *testing.T) {
				check(t, ExpectValidationError(t, publicTS, conf, f, values))
			})
		}
	})
}

func AssertCommonErrorCases(t *testing.T, flows []string) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchemaFromRaw(conf, basicSchema)
	uiTS := testhelpers.NewRegistrationUIFlowEchoServer(t, reg)
	publicTS := setupServer(t, reg)
	apiClient := testhelpers.NewDebugClient(t)
	errTS := testhelpers.NewErrorTestServer(t, reg)

	t.Run("description=can call endpoints only without session", func(t *testing.T) {
		values := url.Values{}
		t.Run("type=browser", func(t *testing.T) {
			res, err := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg).
				Do(httpx.MustNewRequest("POST", publicTS.URL+registration.RouteSubmitFlow, strings.NewReader(values.Encode()), "application/x-www-form-urlencoded"))
			require.NoError(t, err)
			defer res.Body.Close()
			assert.EqualValues(t, http.StatusOK, res.StatusCode, "%+v", res.Request)
			assert.Contains(t, res.Request.URL.String(), conf.Source().String(config.ViperKeySelfServiceBrowserDefaultReturnTo))
		})

		t.Run("type=api", func(t *testing.T) {
			res, err := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg).
				Do(httpx.MustNewRequest("POST", publicTS.URL+registration.RouteSubmitFlow, strings.NewReader(testhelpers.EncodeFormAsJSON(t, true, values)), "application/json"))
			require.NoError(t, err)
			assert.Len(t, res.Cookies(), 0)
			defer res.Body.Close()
			assertx.EqualAsJSON(t, registration.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(ioutilx.MustReadAll(res.Body), "error").Raw))
		})
	})

	t.Run("case=should show the error ui because the request payload is malformed", func(t *testing.T) {
		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)
			body, res := testhelpers.RegistrationMakeRequest(t, true, false, f, apiClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, body, `Expected JSON sent in request body to be an object but got: Number`)
		})

		t.Run("type=spa", func(t *testing.T) {
			f := testhelpers.InitializeRegistrationFlowViaBrowser(t, apiClient, publicTS, true)
			body, res := testhelpers.RegistrationMakeRequest(t, true, false, f, apiClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, body, `Expected JSON sent in request body to be an object but got: Number`)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, false)
			body, res := testhelpers.RegistrationMakeRequest(t, false, false, f, browserClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/registration-ts")
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "invalid URL escape", "%s", body)
			assert.Equal(t, "email", gjson.Get(body, "ui.nodes.#(attributes.name==\"traits.email\").attributes.type").String(), "%s", body)
		})
	})
	t.Run("description=can call endpoints only without session", func(t *testing.T) {
		values := url.Values{}

		t.Run("type=browser", func(t *testing.T) {
			res, err := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg).
				Do(httpx.MustNewRequest("POST", publicTS.URL+registration.RouteSubmitFlow, strings.NewReader(values.Encode()), "application/x-www-form-urlencoded"))
			require.NoError(t, err)
			defer res.Body.Close()
			assert.EqualValues(t, http.StatusOK, res.StatusCode, "%+v", res.Request)
			assert.Contains(t, res.Request.URL.String(), conf.Source().String(config.ViperKeySelfServiceBrowserDefaultReturnTo))
		})

		t.Run("type=api", func(t *testing.T) {
			res, err := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg).
				Do(httpx.MustNewRequest("POST", publicTS.URL+registration.RouteSubmitFlow, strings.NewReader(testhelpers.EncodeFormAsJSON(t, true, values)), "application/json"))
			require.NoError(t, err)
			assert.Len(t, res.Cookies(), 0)
			defer res.Body.Close()
			assertx.EqualAsJSON(t, registration.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(ioutilx.MustReadAll(res.Body), "error").Raw))
		})
	})

	t.Run("case=should show the error ui because the request payload is malformed", func(t *testing.T) {
		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)
			body, res := testhelpers.RegistrationMakeRequest(t, true, false, f, apiClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, body, `Expected JSON sent in request body to be an object but got: Number`)
		})

		t.Run("type=spa", func(t *testing.T) {
			f := testhelpers.InitializeRegistrationFlowViaBrowser(t, apiClient, publicTS, true)
			body, res := testhelpers.RegistrationMakeRequest(t, true, false, f, apiClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, body, `Expected JSON sent in request body to be an object but got: Number`)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, false)
			body, res := testhelpers.RegistrationMakeRequest(t, false, false, f, browserClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/registration-ts")
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "invalid URL escape", "%s", body)
			assert.Equal(t, "email", gjson.Get(body, "ui.nodes.#(attributes.name==\"traits.email\").attributes.type").String(), "%s", body)
		})
	})

	t.Run("case=should show the error ui because the method is missing in payload", func(t *testing.T) {
		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)
			body, res := testhelpers.RegistrationMakeRequest(t, true, false, f, apiClient, "{}}")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "Could not find a strategy to sign you up with. Did you fill out the form correctly?", "%s", body)
		})

		t.Run("type=spa", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, true)
			body, res := testhelpers.RegistrationMakeRequest(t, false, true, f, browserClient, "{}}")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "Could not find a strategy to sign you up with. Did you fill out the form correctly?", "%s", body)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, false)
			body, res := testhelpers.RegistrationMakeRequest(t, false, false, f, browserClient, "foo=bar")
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/registration-ts")
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "Could not find a strategy to sign you up with. Did you fill out the form correctly?", "%s", body)
		})
	})

	t.Run("case=should show the error ui because the request id is missing", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.Equal(t, int64(http.StatusNotFound), gjson.Get(actual, "code").Int(), "%s", actual)
			assert.Equal(t, "Not Found", gjson.Get(actual, "status").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "message").String(), "Unable to locate the resource", "%s", actual)
		}

		fakeFlow := &kratos.SelfServiceRegistrationFlow{Ui: kratos.UiContainer{
			Action: publicTS.URL + registration.RouteSubmitFlow + "?flow=" + x.NewUUID().String(),
		}}

		t.Run("type=api", func(t *testing.T) {
			actual, res := testhelpers.RegistrationMakeRequest(t, true, false, fakeFlow, apiClient, "{}")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
			check(t, gjson.Get(actual, "error").Raw)
		})

		t.Run("type=api", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			actual, res := testhelpers.RegistrationMakeRequest(t, false, true, fakeFlow, browserClient, "{}")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
			check(t, gjson.Get(actual, "error").Raw)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			actual, res := testhelpers.RegistrationMakeRequest(t, false, false, fakeFlow, browserClient, "")
			assert.Contains(t, res.Request.URL.String(), errTS.URL)
			check(t, actual)
		})
	})

	t.Run("case=should return an error because the request is expired", func(t *testing.T) {
		conf.MustSet(config.ViperKeySelfServiceRegistrationRequestLifespan, "500ms")
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeySelfServiceRegistrationRequestLifespan, "10m")
		})

		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)

			time.Sleep(time.Millisecond * 600)
			actual, res := testhelpers.RegistrationMakeRequest(t, true, false, f, apiClient, "{}")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
			assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", gjson.Get(actual, "use_flow_id").String())
			assertx.EqualAsJSONExcept(t, flow.NewFlowExpiredError(time.Now()), json.RawMessage(actual), []string{"use_flow_id", "since"}, "expired", "%s", actual)
		})

		t.Run("type=spa", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, true)

			time.Sleep(time.Millisecond * 600)
			actual, res := testhelpers.RegistrationMakeRequest(t, false, true, f, browserClient, "{}")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
			assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", gjson.Get(actual, "use_flow_id").String())
			assertx.EqualAsJSONExcept(t, flow.NewFlowExpiredError(time.Now()), json.RawMessage(actual), []string{"use_flow_id", "since"}, "expired", "%s", actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, false)

			time.Sleep(time.Millisecond * 600)
			actual, res := testhelpers.RegistrationMakeRequest(t, false, false, f, browserClient, "")
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/registration-ts")
			assert.NotEqual(t, f.Id, gjson.Get(actual, "id").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "ui.messages.0.text").String(), "expired", "%s", actual)
		})
	})

	t.Run("case=should fail because the return_to url is not allowed", func(t *testing.T) {
		testhelpers.SetDefaultIdentitySchemaFromRaw(conf, multifieldSchema)
		t.Cleanup(func() {
			testhelpers.SetDefaultIdentitySchemaFromRaw(conf, basicSchema)
		})

		email := testhelpers.RandomEmail()
		var check = func(t *testing.T, actual string) {
			assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "ui.action").String(), publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
			CheckFormContent(t, []byte(actual), "password", "csrf_token", "traits.username", "traits.foobar")
			assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==password).messages.0").String(), "data breaches and must no longer be used.", "%s", actual)

			// but the method is still set
			assert.Equal(t, "password", gjson.Get(actual, "ui.nodes.#(attributes.name==method).attributes.value").String(), "%s", actual)
		}

		var values = func(v url.Values) {
			v.Set("traits.username", email)
			v.Set("password", "password")
			v.Set("traits.foobar", "bar")
		}

		for _, f := range flows {
			t.Run("type="+f, func(t *testing.T) {
				check(t, ExpectValidationError(t, publicTS, conf, f, values))
			})
		}
	})
}
