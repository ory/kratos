package registration_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/ory/kratos-client-go"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"
	"github.com/ory/x/urlx"

	"github.com/ory/herodot"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func TestHandleError(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/login.schema.json")

	public, admin := testhelpers.NewKratosServer(t, reg)

	router := httprouter.New()
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	testhelpers.NewRegistrationUIFlowEchoServer(t, reg)
	testhelpers.NewErrorTestServer(t, reg)

	h := reg.RegistrationFlowErrorHandler()
	sdk := testhelpers.NewSDKClient(admin)

	var registrationFlow *registration.Flow
	var flowError error
	var ct identity.CredentialsType
	router.GET("/error", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.WriteFlowError(w, r, ct, registrationFlow, flowError)
	})

	reset := func() {
		registrationFlow = nil
		flowError = nil
		ct = ""
	}

	newFlow := func(t *testing.T, ttl time.Duration, ft flow.Type) *registration.Flow {
		req := &http.Request{URL: urlx.ParseOrPanic("/")}
		f := registration.NewFlow(ttl, "csrf_token", req, ft)
		for _, s := range reg.RegistrationStrategies(context.Background()) {
			require.NoError(t, s.PopulateRegistrationMethod(req, f))
		}

		require.NoError(t, reg.RegistrationFlowPersister().CreateRegistrationFlow(context.Background(), f))
		return f
	}

	expectErrorUI := func(t *testing.T) (interface{}, *http.Response) {
		res, err := ts.Client().Get(ts.URL + "/error")
		require.NoError(t, err)
		defer res.Body.Close()
		require.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowErrorURL().String()+"?error=")

		sse, _, err := sdk.PublicApi.GetSelfServiceError(context.Background()).Error_(res.Request.URL.Query().Get("error")).Execute()
		require.NoError(t, err)

		return sse.Errors, nil
	}

	anHourAgo := time.Now().Add(-time.Hour)

	t.Run("case=error with nil flow defaults to error ui redirect", func(t *testing.T) {
		t.Cleanup(reset)

		flowError = herodot.ErrInternalServerError.WithReason("system error")
		ct = identity.CredentialsTypePassword

		sse, _ := expectErrorUI(t)
		assertx.EqualAsJSON(t, []interface{}{flowError}, sse)
	})

	t.Run("case=error with nil flow detects application/json", func(t *testing.T) {
		t.Cleanup(reset)

		flowError = herodot.ErrInternalServerError.WithReason("system error")
		ct = identity.CredentialsTypePassword

		res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
		require.NoError(t, err)
		defer res.Body.Close()
		assert.Contains(t, res.Header.Get("Content-Type"), "application/json")
		assert.NotContains(t, res.Request.URL.String(), conf.SelfServiceFlowErrorURL().String()+"?error=")

		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "system error")
	})

	t.Run("flow=api", func(t *testing.T) {
		t.Run("case=expired error", func(t *testing.T) {
			t.Cleanup(reset)

			registrationFlow = newFlow(t, time.Minute, flow.TypeAPI)
			flowError = registration.NewFlowExpiredError(anHourAgo)
			ct = identity.CredentialsTypePassword

			res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
			require.NoError(t, err)
			defer res.Body.Close()
			require.Contains(t, res.Request.URL.String(), public.URL+registration.RouteGetFlow)
			require.Equal(t, http.StatusOK, res.StatusCode)

			body, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, int(text.ErrorValidationRegistrationFlowExpired), int(gjson.GetBytes(body, "messages.0.id").Int()))
			assert.NotEqual(t, registrationFlow.ID.String(), gjson.GetBytes(body, "id").String())
		})

		t.Run("case=validation error", func(t *testing.T) {
			t.Cleanup(reset)

			registrationFlow = newFlow(t, time.Minute, flow.TypeAPI)
			flowError = schema.NewInvalidCredentialsError()
			ct = identity.CredentialsTypePassword

			res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
			require.NoError(t, err)
			defer res.Body.Close()
			require.Equal(t, http.StatusBadRequest, res.StatusCode)

			body, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(gjson.GetBytes(body, "methods.password.config.messages.0.id").Int()), "%s", body)
			assert.Equal(t, registrationFlow.ID.String(), gjson.GetBytes(body, "id").String())
		})

		t.Run("case=generic error", func(t *testing.T) {
			t.Cleanup(reset)

			registrationFlow = newFlow(t, time.Minute, flow.TypeAPI)
			flowError = herodot.ErrInternalServerError.WithReason("system error")
			ct = identity.CredentialsTypePassword

			res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
			require.NoError(t, err)
			defer res.Body.Close()
			require.Equal(t, http.StatusInternalServerError, res.StatusCode)

			body, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			assert.JSONEq(t, x.MustEncodeJSON(t, flowError), gjson.GetBytes(body, "error").Raw)
		})

		t.Run("case=method is unknown", func(t *testing.T) {
			t.Cleanup(reset)

			registrationFlow = newFlow(t, time.Minute, flow.TypeAPI)
			flowError = herodot.ErrInternalServerError.WithReason("system error")
			ct = "invalid-method"

			res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
			require.NoError(t, err)
			defer res.Body.Close()
			require.Equal(t, http.StatusInternalServerError, res.StatusCode)

			body, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Contains(t, gjson.GetBytes(body, "error.message").String(), "invalid-method", "%s", body)
		})
	})

	t.Run("flow=browser", func(t *testing.T) {
		expectRegistrationUI := func(t *testing.T) (*kratos.RegistrationFlow, *http.Response) {
			res, err := ts.Client().Get(ts.URL + "/error")
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRegistrationUI().String()+"?flow=")

			lf, _, err := sdk.PublicApi.GetSelfServiceRegistrationFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
			require.NoError(t, err)
			return lf, res
		}

		t.Run("case=expired error", func(t *testing.T) {
			t.Cleanup(reset)

			registrationFlow = &registration.Flow{Type: flow.TypeBrowser}
			flowError = registration.NewFlowExpiredError(anHourAgo)
			ct = identity.CredentialsTypePassword

			lf, _ := expectRegistrationUI(t)
			require.Len(t, lf.Messages, 1)
			assert.Equal(t, int(text.ErrorValidationRegistrationFlowExpired), int(lf.Messages[0].Id))
		})

		t.Run("case=validation error", func(t *testing.T) {
			t.Cleanup(reset)

			registrationFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = schema.NewInvalidCredentialsError()
			ct = identity.CredentialsTypePassword

			lf, _ := expectRegistrationUI(t)
			require.NotEmpty(t, lf.Methods[string(ct)], x.MustEncodeJSON(t, lf))
			require.Len(t, lf.Methods[string(ct)].Config.Messages, 1, x.MustEncodeJSON(t, lf))
			assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(lf.Methods[string(ct)].Config.Messages[0].Id), x.MustEncodeJSON(t, lf))
		})

		t.Run("case=generic error", func(t *testing.T) {
			t.Cleanup(reset)

			registrationFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = herodot.ErrInternalServerError.WithReason("system error")
			ct = identity.CredentialsTypePassword

			sse, _ := expectErrorUI(t)
			assertx.EqualAsJSON(t, []interface{}{flowError}, sse)
		})

		t.Run("case=method is unknown", func(t *testing.T) {
			t.Cleanup(reset)

			registrationFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = herodot.ErrInternalServerError.WithReason("system error")
			ct = "invalid-method"

			sse, _ := expectErrorUI(t)
			body := x.MustEncodeJSON(t, sse)
			assert.Contains(t, gjson.Get(body, "0.message").String(), "invalid-method", "%s", body)
		})
	})
}
