package recovery_test

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

	sdkp "github.com/ory/kratos-client-go/client/public"
	"github.com/ory/kratos-client-go/models"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func TestHandleError(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceRecoveryEnabled, true)

	public, admin := testhelpers.NewKratosServer(t, reg)

	router := httprouter.New()
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	testhelpers.NewErrorTestServer(t, reg)

	h := reg.RecoveryFlowErrorHandler()
	sdk := testhelpers.NewSDKClient(admin)

	var recoveryFlow *recovery.Flow
	var flowError error
	var methodName string
	router.GET("/error", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.WriteFlowError(w, r, methodName, recoveryFlow, flowError)
	})

	reset := func() {
		recoveryFlow = nil
		flowError = nil
		methodName = ""
	}

	newFlow := func(t *testing.T, ttl time.Duration, ft flow.Type) *recovery.Flow {
		req := &http.Request{URL: urlx.ParseOrPanic("/")}
		f, err := recovery.NewFlow(ttl, x.FakeCSRFToken, req, reg.RecoveryStrategies(), ft)
		require.NoError(t, err)
		require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(context.Background(), f))
		f, err = reg.RecoveryFlowPersister().GetRecoveryFlow(context.Background(), f.ID)
		require.NoError(t, err)
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
		methodName = recovery.StrategyRecoveryLinkName

		sse, _ := expectErrorUI(t)
		assertx.EqualAsJSON(t, []interface{}{flowError}, sse)
	})

	t.Run("case=error with nil flow detects application/json", func(t *testing.T) {
		t.Cleanup(reset)

		flowError = herodot.ErrInternalServerError.WithReason("system error")
		methodName = recovery.StrategyRecoveryLinkName

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

			recoveryFlow = newFlow(t, time.Minute, flow.TypeAPI)
			flowError = recovery.NewFlowExpiredError(anHourAgo)
			methodName = recovery.StrategyRecoveryLinkName

			res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
			require.NoError(t, err)
			defer res.Body.Close()
			require.Contains(t, res.Request.URL.String(), public.URL+recovery.RouteGetFlow)
			require.Equal(t, http.StatusOK, res.StatusCode, "%+v", res.Request)

			body, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, int(text.ErrorValidationRecoveryFlowExpired), int(gjson.GetBytes(body, "messages.0.id").Int()))
			assert.NotEqual(t, recoveryFlow.ID.String(), gjson.GetBytes(body, "id").String())
		})

		t.Run("case=validation error", func(t *testing.T) {
			t.Cleanup(reset)

			recoveryFlow = newFlow(t, time.Minute, flow.TypeAPI)
			flowError = schema.NewInvalidCredentialsError()
			methodName = recovery.StrategyRecoveryLinkName

			res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
			require.NoError(t, err)
			defer res.Body.Close()
			require.Equal(t, http.StatusBadRequest, res.StatusCode)

			body, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(gjson.GetBytes(body, "methods.link.config.messages.0.id").Int()), "%s", body)
			assert.Equal(t, recoveryFlow.ID.String(), gjson.GetBytes(body, "id").String())
		})

		t.Run("case=generic error", func(t *testing.T) {
			t.Cleanup(reset)

			recoveryFlow = newFlow(t, time.Minute, flow.TypeAPI)
			flowError = herodot.ErrInternalServerError.WithReason("system error")
			methodName = recovery.StrategyRecoveryLinkName

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

			recoveryFlow = newFlow(t, time.Minute, flow.TypeAPI)
			flowError = herodot.ErrInternalServerError.WithReason("system error")
			methodName = "invalid-method"

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
		expectRecoveryUI := func(t *testing.T) (*models.RecoveryFlow, *http.Response) {
			res, err := ts.Client().Get(ts.URL + "/error")
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String()+"?flow=")

			lf, err := sdk.Public.GetSelfServiceRecoveryFlow(sdkp.NewGetSelfServiceRecoveryFlowParams().
				WithID(res.Request.URL.Query().Get("flow")))
			require.NoError(t, err)
			return lf.Payload, res
		}

		t.Run("case=expired error", func(t *testing.T) {
			t.Cleanup(reset)

			recoveryFlow = &recovery.Flow{Type: flow.TypeBrowser}
			flowError = recovery.NewFlowExpiredError(anHourAgo)
			methodName = recovery.StrategyRecoveryLinkName

			lf, _ := expectRecoveryUI(t)
			require.Len(t, lf.Messages, 1)
			assert.Equal(t, int(text.ErrorValidationRecoveryFlowExpired), int(lf.Messages[0].ID))
		})

		t.Run("case=validation error", func(t *testing.T) {
			t.Cleanup(reset)

			recoveryFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = schema.NewInvalidCredentialsError()
			methodName = recovery.StrategyRecoveryLinkName

			lf, _ := expectRecoveryUI(t)
			require.NotEmpty(t, lf.Methods[string(methodName)], x.MustEncodeJSON(t, lf))
			require.Len(t, lf.Methods[string(methodName)].Config.Messages, 1, x.MustEncodeJSON(t, lf))
			assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(lf.Methods[string(methodName)].Config.Messages[0].ID), x.MustEncodeJSON(t, lf))
		})

		t.Run("case=generic error", func(t *testing.T) {
			t.Cleanup(reset)

			recoveryFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = herodot.ErrInternalServerError.WithReason("system error")
			methodName = recovery.StrategyRecoveryLinkName

			sse, _ := expectErrorUI(t)
			assertx.EqualAsJSON(t, []interface{}{flowError}, sse)
		})

		t.Run("case=method is unknown", func(t *testing.T) {
			t.Cleanup(reset)

			recoveryFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = herodot.ErrInternalServerError.WithReason("system error")
			methodName = "invalid-method"

			sse, _ := expectErrorUI(t)
			body := x.MustEncodeJSON(t, sse)
			assert.Contains(t, gjson.Get(body, "0.message").String(), "invalid-method", "%s", body)
		})
	})
}
