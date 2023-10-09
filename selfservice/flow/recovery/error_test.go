// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package recovery_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/x/jsonx"
	"github.com/ory/x/snapshotx"

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
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func TestHandleError(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryEnabled, true)
	conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryUse, "code")

	public, _ := testhelpers.NewKratosServer(t, reg)

	router := httprouter.New()
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	testhelpers.NewErrorTestServer(t, reg)

	h := reg.RecoveryFlowErrorHandler()
	sdk := testhelpers.NewSDKClient(public)

	var recoveryFlow *recovery.Flow
	var flowError error
	var methodName node.UiNodeGroup
	router.GET("/error", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.WriteFlowError(w, r, recoveryFlow, methodName, flowError)
	})

	reset := func() {
		recoveryFlow = nil
		flowError = nil
		methodName = ""
	}

	newFlow := func(t *testing.T, ttl time.Duration, ft flow.Type) *recovery.Flow {
		req := &http.Request{URL: urlx.ParseOrPanic("/")}
		s, err := reg.GetActiveRecoveryStrategy(context.Background())
		require.NoError(t, err)
		f, err := recovery.NewFlow(conf, ttl, x.FakeCSRFToken, req, s, ft)
		require.NoError(t, err)
		require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(context.Background(), f))
		f, err = reg.RecoveryFlowPersister().GetRecoveryFlow(context.Background(), f.ID)
		require.NoError(t, err)
		return f
	}

	expectErrorUI := func(t *testing.T) (map[string]interface{}, *http.Response) {
		res, err := ts.Client().Get(ts.URL + "/error")
		require.NoError(t, err)
		defer res.Body.Close()
		require.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowErrorURL(ctx).String()+"?id=")

		sse, _, err := sdk.FrontendApi.GetFlowError(context.Background()).Id(res.Request.URL.Query().Get("id")).Execute()
		require.NoError(t, err)

		return sse.Error, nil
	}

	anHourAgo := time.Now().Add(-time.Hour)

	t.Run("case=error with nil flow defaults to error ui redirect", func(t *testing.T) {
		t.Cleanup(reset)

		flowError = herodot.ErrInternalServerError.WithReason("system error")
		methodName = node.UiNodeGroup(recovery.RecoveryStrategyLink)

		sse, _ := expectErrorUI(t)
		assertx.EqualAsJSON(t, flowError, sse)
	})

	t.Run("case=error with nil flow detects application/json", func(t *testing.T) {
		t.Cleanup(reset)

		flowError = herodot.ErrInternalServerError.WithReason("system error")
		methodName = node.UiNodeGroup(recovery.RecoveryStrategyLink)

		res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
		require.NoError(t, err)
		defer res.Body.Close()
		assert.Contains(t, res.Header.Get("Content-Type"), "application/json")
		assert.NotContains(t, res.Request.URL.String(), conf.SelfServiceFlowErrorURL(ctx).String()+"?id=")

		body, err := io.ReadAll(res.Body)
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

				recoveryFlow = newFlow(t, time.Minute, flow.TypeAPI)
				flowError = flow.NewFlowExpiredError(anHourAgo)
				methodName = node.UiNodeGroup(recovery.RecoveryStrategyLink)

				res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()
				require.Contains(t, res.Request.URL.String(), public.URL+recovery.RouteGetFlow)
				require.Equal(t, http.StatusOK, res.StatusCode, "%+v", res.Request)

				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, int(text.ErrorValidationRecoveryFlowExpired), int(gjson.GetBytes(body, "ui.messages.0.id").Int()), string(body))
				assert.NotEqual(t, recoveryFlow.ID.String(), gjson.GetBytes(body, "id").String())
			})

			t.Run("case=validation error", func(t *testing.T) {
				t.Cleanup(reset)

				recoveryFlow = newFlow(t, time.Minute, tc.t)
				flowError = schema.NewInvalidCredentialsError()
				methodName = node.UiNodeGroup(recovery.RecoveryStrategyLink)

				res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()
				require.Equal(t, http.StatusBadRequest, res.StatusCode)

				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(gjson.GetBytes(body, "ui.messages.0.id").Int()), "%s", body)
				assert.Equal(t, recoveryFlow.ID.String(), gjson.GetBytes(body, "id").String())
			})

			t.Run("case=generic error", func(t *testing.T) {
				t.Cleanup(reset)

				recoveryFlow = newFlow(t, time.Minute, tc.t)
				flowError = herodot.ErrInternalServerError.WithReason("system error")
				methodName = node.UiNodeGroup(recovery.RecoveryStrategyLink)

				res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()
				require.Equal(t, http.StatusInternalServerError, res.StatusCode)

				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.JSONEq(t, x.MustEncodeJSON(t, flowError), gjson.GetBytes(body, "error").Raw)
			})

			t.Run("case=fails if active strategy is disabled", func(t *testing.T) {
				c, reg := internal.NewVeryFastRegistryWithoutDB(t)
				c.Set(context.Background(), "selfservice.methods.code.enabled", false)
				c.Set(context.Background(), config.ViperKeySelfServiceRecoveryUse, "code")
				_, err := reg.GetActiveRecoveryStrategy(context.Background())
				recoveryFlow = newFlow(t, time.Minute, tc.t)
				flowError = err
				methodName = node.UiNodeGroup(recovery.RecoveryStrategyLink)

				res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()
				require.Equal(t, http.StatusBadRequest, res.StatusCode)

				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				snapshotx.SnapshotTJSON(t, body, snapshotx.ExceptPaths("id", "expires_at", "issued_at", "ui.action", "ui.nodes.0.attributes.value"))
			})
		})
	}

	t.Run("flow=browser", func(t *testing.T) {
		expectRecoveryUI := func(t *testing.T) (*recovery.Flow, *http.Response) {
			res, err := ts.Client().Get(ts.URL + "/error")
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI(ctx).String()+"?flow=")

			rf, err := reg.RecoveryFlowPersister().GetRecoveryFlow(context.Background(), uuid.FromStringOrNil(res.Request.URL.Query().Get("flow")))
			require.NoError(t, err)
			return rf, res
		}

		t.Run("case=expired error", func(t *testing.T) {
			t.Cleanup(reset)

			recoveryFlow = &recovery.Flow{Type: flow.TypeBrowser}
			flowError = flow.NewFlowExpiredError(anHourAgo)
			methodName = node.LinkGroup

			lf, _ := expectRecoveryUI(t)
			require.Len(t, lf.UI.Messages, 1, "%s", jsonx.TestMarshalJSONString(t, lf))
			assert.Equal(t, int(text.ErrorValidationRecoveryFlowExpired), int(lf.UI.Messages[0].ID))
		})

		t.Run("case=validation error", func(t *testing.T) {
			t.Cleanup(reset)

			recoveryFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = schema.NewInvalidCredentialsError()
			methodName = node.LinkGroup

			lf, _ := expectRecoveryUI(t)
			require.NotEmpty(t, lf.UI, x.MustEncodeJSON(t, lf))
			require.Len(t, lf.UI.Messages, 1, x.MustEncodeJSON(t, lf))
			assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(lf.UI.Messages[0].ID), x.MustEncodeJSON(t, lf))
		})

		t.Run("case=generic error", func(t *testing.T) {
			t.Cleanup(reset)

			recoveryFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = herodot.ErrInternalServerError.WithReason("system error")
			methodName = node.LinkGroup

			sse, _ := expectErrorUI(t)
			assertx.EqualAsJSON(t, flowError, sse)
		})

		t.Run("case=new flow uses strategy of old flow", func(t *testing.T) {

			t.Cleanup(reset)

			recoveryFlow = &recovery.Flow{Type: flow.TypeBrowser, Active: "code"}
			flowError = flow.NewFlowExpiredError(anHourAgo)

			lf, _ := expectRecoveryUI(t)
			require.Len(t, lf.UI.Messages, 1, "%s", jsonx.TestMarshalJSONString(t, lf))
			assert.Equal(t, int(text.ErrorValidationRecoveryFlowExpired), int(lf.UI.Messages[0].ID))
			assert.Equal(t, recoveryFlow.Active.String(), lf.Active.String())
		})

		t.Run("case=new flow uses current strategy if strategy of old flow does not exist", func(t *testing.T) {

			t.Cleanup(reset)

			recoveryFlow = &recovery.Flow{Type: flow.TypeBrowser, Active: "not-valid"}
			flowError = flow.NewFlowExpiredError(anHourAgo)

			lf, _ := expectRecoveryUI(t)
			require.Len(t, lf.UI.Messages, 1, "%s", jsonx.TestMarshalJSONString(t, lf))
			assert.Equal(t, int(text.ErrorValidationRecoveryFlowExpired), int(lf.UI.Messages[0].ID))
			assert.Equal(t, "code", lf.Active.String())
		})

		t.Run("case=fails to retry flow if recovery strategy id is not valid", func(t *testing.T) {

			t.Cleanup(func() {
				reset()
				conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryUse, "code")
			})

			recoveryFlow = newFlow(t, 0, flow.TypeBrowser)
			recoveryFlow.Active = "not-valid"
			flowError = flow.NewFlowExpiredError(anHourAgo)

			conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryUse, "not-valid")
			sse, _ := expectErrorUI(t)
			snapshotx.SnapshotT(t, sse)
		})
	})
}
