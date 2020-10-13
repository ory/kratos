package login_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"
	"github.com/ory/x/urlx"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	sdkp "github.com/ory/kratos/internal/httpclient/client/public"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func TestHandleError(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	public, admin := testhelpers.NewKratosServer(t, reg)

	router := httprouter.New()
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	testhelpers.NewLoginUIFlowEchoServer(t, reg)
	testhelpers.NewErrorTestServer(t, reg)

	h := reg.LoginFlowErrorHandler()
	sdk := testhelpers.NewSDKClient(admin)

	var loginFlow *login.Flow
	var flowError error
	var ct identity.CredentialsType
	router.GET("/error", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.WriteFlowError(w, r, ct, loginFlow, flowError)
	})

	reset := func() {
		loginFlow = nil
		flowError = nil
		ct = ""
	}

	newFlow := func(t *testing.T, ttl time.Duration, ft flow.Type) *login.Flow {
		req := &http.Request{URL: urlx.ParseOrPanic("/")}
		f := login.NewFlow(ttl, "csrf_token", req, ft)
		for _, s := range reg.LoginStrategies() {
			require.NoError(t, s.PopulateLoginMethod(req, f))
		}

		require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), f))
		return f
	}

	expectErrorUI := func(t *testing.T) (interface{}, *http.Response) {
		res, err := ts.Client().Get(ts.URL + "/error")
		require.NoError(t, err)
		defer res.Body.Close()
		require.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowErrorURL().String()+"?error=")

		sse, err := sdk.Public.GetSelfServiceError(sdkp.NewGetSelfServiceErrorParams().
			WithError(res.Request.URL.Query().Get("error")))
		require.NoError(t, err)

		return sse.Payload.Errors, nil
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

			loginFlow = newFlow(t, time.Minute, flow.TypeAPI)
			flowError = login.NewFlowExpiredError(anHourAgo)
			ct = identity.CredentialsTypePassword

			res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
			require.NoError(t, err)
			defer res.Body.Close()
			require.Contains(t, res.Request.URL.String(), public.URL+login.RouteGetFlow)
			require.Equal(t, http.StatusOK, res.StatusCode)

			body, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, int(text.ErrorValidationLoginFlowExpired), int(gjson.GetBytes(body, "messages.0.id").Int()))
			assert.NotEqual(t, loginFlow.ID.String(), gjson.GetBytes(body, "id").String())
		})

		t.Run("case=validation error", func(t *testing.T) {
			t.Cleanup(reset)

			loginFlow = newFlow(t, time.Minute, flow.TypeAPI)
			flowError = schema.NewInvalidCredentialsError()
			ct = identity.CredentialsTypePassword

			res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
			require.NoError(t, err)
			defer res.Body.Close()
			require.Equal(t, http.StatusBadRequest, res.StatusCode)

			body, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(gjson.GetBytes(body, "methods.password.config.messages.0.id").Int()), "%s", body)
			assert.Equal(t, loginFlow.ID.String(), gjson.GetBytes(body, "id").String())
		})

		t.Run("case=generic error", func(t *testing.T) {
			t.Cleanup(reset)

			loginFlow = newFlow(t, time.Minute, flow.TypeAPI)
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

			loginFlow = newFlow(t, time.Minute, flow.TypeAPI)
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
		expectLoginUI := func(t *testing.T) (*models.LoginFlow, *http.Response) {
			res, err := ts.Client().Get(ts.URL + "/error")
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowLoginUI().String()+"?flow=")

			lf, err := sdk.Public.GetSelfServiceLoginFlow(sdkp.NewGetSelfServiceLoginFlowParams().
				WithID(res.Request.URL.Query().Get("flow")))
			require.NoError(t, err)
			return lf.Payload, res
		}

		t.Run("case=expired error", func(t *testing.T) {
			t.Cleanup(reset)

			loginFlow = &login.Flow{Type: flow.TypeBrowser}
			flowError = login.NewFlowExpiredError(anHourAgo)
			ct = identity.CredentialsTypePassword

			lf, _ := expectLoginUI(t)
			require.Len(t, lf.Messages, 1)
			assert.Equal(t, int(text.ErrorValidationLoginFlowExpired), int(lf.Messages[0].ID))
		})

		t.Run("case=validation error", func(t *testing.T) {
			t.Cleanup(reset)

			loginFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = schema.NewInvalidCredentialsError()
			ct = identity.CredentialsTypePassword

			lf, _ := expectLoginUI(t)
			require.NotEmpty(t, lf.Methods[string(ct)], x.MustEncodeJSON(t, lf))
			require.Len(t, lf.Methods[string(ct)].Config.Messages, 1, x.MustEncodeJSON(t, lf))
			assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(lf.Methods[string(ct)].Config.Messages[0].ID), x.MustEncodeJSON(t, lf))
		})

		t.Run("case=generic error", func(t *testing.T) {
			t.Cleanup(reset)

			loginFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = herodot.ErrInternalServerError.WithReason("system error")
			ct = identity.CredentialsTypePassword

			sse, _ := expectErrorUI(t)
			assertx.EqualAsJSON(t, []interface{}{flowError}, sse)
		})

		t.Run("case=method is unknown", func(t *testing.T) {
			t.Cleanup(reset)

			loginFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = herodot.ErrInternalServerError.WithReason("system error")
			ct = "invalid-method"

			sse, _ := expectErrorUI(t)
			body := x.MustEncodeJSON(t, sse)
			assert.Contains(t, gjson.Get(body, "0.message").String(), "invalid-method", "%s", body)
		})
	})
}
