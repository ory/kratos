package verification_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/ory/kratos/schema/errors"

	"github.com/gofrs/uuid"

	"github.com/ory/x/jsonx"

	"github.com/ory/kratos/ui/node"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"
	"github.com/ory/x/urlx"

	"github.com/ory/herodot"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func TestHandleError(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceVerificationEnabled, true)

	public, admin := testhelpers.NewKratosServer(t, reg)

	router := httprouter.New()
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	testhelpers.NewVerificationUIFlowEchoServer(t, reg)
	testhelpers.NewErrorTestServer(t, reg)

	h := reg.VerificationFlowErrorHandler()
	sdk := testhelpers.NewSDKClient(admin)

	var verificationFlow *verification.Flow
	var flowError error
	var methodName node.Group
	router.GET("/error", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.WriteFlowError(w, r, verificationFlow, methodName, flowError)
	})

	reset := func() {
		verificationFlow = nil
		flowError = nil
		methodName = ""
	}

	newFlow := func(t *testing.T, ttl time.Duration, ft flow.Type) *verification.Flow {
		req := &http.Request{URL: urlx.ParseOrPanic("/")}
		f, err := verification.NewFlow(conf, ttl, x.FakeCSRFToken, req, reg.VerificationStrategies(context.Background()), ft)
		require.NoError(t, err)
		require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(context.Background(), f))
		f, err = reg.VerificationFlowPersister().GetVerificationFlow(context.Background(), f.ID)
		require.NoError(t, err)
		return f
	}

	expectErrorUI := func(t *testing.T) (map[string]interface{}, *http.Response) {
		res, err := ts.Client().Get(ts.URL + "/error")
		require.NoError(t, err)
		defer res.Body.Close()
		require.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowErrorURL().String()+"?id=")

		sse, _, err := sdk.V0alpha2Api.GetSelfServiceError(context.Background()).Id(res.Request.URL.Query().Get("id")).Execute()
		require.NoError(t, err)

		return sse.Error, nil
	}

	anHourAgo := time.Now().Add(-time.Hour)

	t.Run("case=error with nil flow defaults to error ui redirect", func(t *testing.T) {
		t.Cleanup(reset)

		flowError = herodot.ErrInternalServerError.WithReason("system error")
		methodName = verification.StrategyVerificationLinkName

		sse, _ := expectErrorUI(t)
		assertx.EqualAsJSON(t, flowError, sse)
	})

	t.Run("case=error with nil flow detects application/json", func(t *testing.T) {
		t.Cleanup(reset)

		flowError = herodot.ErrInternalServerError.WithReason("system error")
		methodName = verification.StrategyVerificationLinkName

		res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
		require.NoError(t, err)
		defer res.Body.Close()
		assert.Contains(t, res.Header.Get("Content-Type"), "application/json")
		assert.NotContains(t, res.Request.URL.String(), conf.SelfServiceFlowErrorURL().String()+"?id=")

		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "system error")
	})

	for _, tc := range []struct {
		n string
		t flow.Type
	}{
		{"api", flow.TypeAPI},
		{"spa", flow.TypeBrowser},
	} {
		t.Run("flow="+tc.n, func(t *testing.T) {
			t.Run("case=expired error", func(t *testing.T) {
				t.Cleanup(reset)

				verificationFlow = newFlow(t, time.Minute, flow.TypeAPI)
				flowError = flow.NewFlowExpiredError(anHourAgo)
				methodName = verification.StrategyVerificationLinkName

				res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()
				require.Contains(t, res.Request.URL.String(), public.URL+verification.RouteGetFlow)
				require.Equal(t, http.StatusOK, res.StatusCode, "%+v", res.Request)

				body, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, int(text.ErrorValidationVerificationFlowExpired), int(gjson.GetBytes(body, "ui.messages.0.id").Int()), string(body))
				assert.NotEqual(t, verificationFlow.ID.String(), gjson.GetBytes(body, "id").String())
			})

			t.Run("case=validation error", func(t *testing.T) {
				t.Cleanup(reset)

				verificationFlow = newFlow(t, time.Minute, tc.t)
				flowError = errors.NewInvalidCredentialsError()
				methodName = verification.StrategyVerificationLinkName

				res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()
				require.Equal(t, http.StatusBadRequest, res.StatusCode)

				body, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(gjson.GetBytes(body, "ui.messages.0.id").Int()), "%s", body)
				assert.Equal(t, verificationFlow.ID.String(), gjson.GetBytes(body, "id").String())
			})

			t.Run("case=generic error", func(t *testing.T) {
				t.Cleanup(reset)

				verificationFlow = newFlow(t, time.Minute, tc.t)
				flowError = herodot.ErrInternalServerError.WithReason("system error")
				methodName = verification.StrategyVerificationLinkName

				res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()
				require.Equal(t, http.StatusInternalServerError, res.StatusCode)

				body, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				assert.JSONEq(t, x.MustEncodeJSON(t, flowError), gjson.GetBytes(body, "error").Raw)
			})
		})
	}

	t.Run("flow=browser", func(t *testing.T) {
		expectVerificationUI := func(t *testing.T) (*verification.Flow, *http.Response) {
			res, err := ts.Client().Get(ts.URL + "/error")
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI().String()+"?flow=")

			vf, err := reg.VerificationFlowPersister().GetVerificationFlow(context.Background(), uuid.FromStringOrNil(res.Request.URL.Query().Get("flow")))
			require.NoError(t, err)
			return vf, res
		}

		t.Run("case=expired error", func(t *testing.T) {
			t.Cleanup(reset)

			verificationFlow = &verification.Flow{Type: flow.TypeBrowser}
			flowError = flow.NewFlowExpiredError(anHourAgo)
			methodName = node.VerificationLinkGroup

			lf, _ := expectVerificationUI(t)
			require.Len(t, lf.UI.Messages, 1, "%s", jsonx.TestMarshalJSONString(t, lf))
			assert.Equal(t, int(text.ErrorValidationVerificationFlowExpired), int(lf.UI.Messages[0].ID))
		})

		t.Run("case=validation error", func(t *testing.T) {
			t.Cleanup(reset)

			verificationFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = errors.NewInvalidCredentialsError()
			methodName = node.VerificationLinkGroup

			lf, _ := expectVerificationUI(t)
			require.NotEmpty(t, lf.UI, x.MustEncodeJSON(t, lf))
			require.Len(t, lf.UI.Messages, 1, x.MustEncodeJSON(t, lf))
			assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(lf.UI.Messages[0].ID), x.MustEncodeJSON(t, lf))
		})

		t.Run("case=generic error", func(t *testing.T) {
			t.Cleanup(reset)

			verificationFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = herodot.ErrInternalServerError.WithReason("system error")
			methodName = node.VerificationLinkGroup

			sse, _ := expectErrorUI(t)
			assertx.EqualAsJSON(t, flowError, sse)
		})
	})
}
