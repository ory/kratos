package login_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/ory/kratos/schema/errors"

	"github.com/ory/kratos/identity"

	"github.com/gofrs/uuid"

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
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func TestHandleError(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	_, admin := testhelpers.NewKratosServer(t, reg)

	router := httprouter.New()
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	testhelpers.NewLoginUIFlowEchoServer(t, reg)
	testhelpers.NewErrorTestServer(t, reg)

	h := reg.LoginFlowErrorHandler()
	sdk := testhelpers.NewSDKClient(admin)

	var loginFlow *login.Flow
	var flowError error
	var ct node.Group
	router.GET("/error", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.WriteFlowError(w, r, loginFlow, ct, flowError)
	})

	reset := func() {
		loginFlow = nil
		flowError = nil
		ct = ""
	}

	newFlow := func(t *testing.T, ttl time.Duration, ft flow.Type) *login.Flow {
		req := &http.Request{URL: urlx.ParseOrPanic("/")}
		f, err := login.NewFlow(conf, ttl, "csrf_token", req, ft)
		require.NoError(t, err)

		for _, s := range reg.LoginStrategies(context.Background()) {
			require.NoError(t, s.PopulateLoginMethod(req, identity.AuthenticatorAssuranceLevel1, f))
		}

		require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), f))
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
		ct = node.PasswordGroup

		sse, _ := expectErrorUI(t)
		assertx.EqualAsJSON(t, flowError, sse)
	})

	t.Run("case=relative error", func(t *testing.T) {
		t.Cleanup(reset)
		reg.Config(context.Background()).MustSet(config.ViperKeySelfServiceErrorUI, "/login-ts")
		flowError = herodot.ErrInternalServerError.WithReason("system error")
		ct = node.PasswordGroup
		assert.Regexp(
			t,
			"^/login-ts.*$",
			testhelpers.GetSelfServiceRedirectLocation(t, ts.URL+"/error"),
		)

	})

	t.Run("case=error with nil flow detects application/json", func(t *testing.T) {
		t.Cleanup(reset)

		flowError = herodot.ErrInternalServerError.WithReason("system error")
		ct = node.PasswordGroup

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

				loginFlow = newFlow(t, time.Minute, tc.t)
				flowError = flow.NewFlowExpiredError(anHourAgo)
				ct = node.PasswordGroup

				res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()

				body, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				require.Equal(t, http.StatusGone, res.StatusCode, "%+v\n\t%s", res.Request, body)

				assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", gjson.GetBytes(body, "use_flow_id").String())
				assertx.EqualAsJSONExcept(t, flow.NewFlowExpiredError(anHourAgo), json.RawMessage(body), []string{"since", "redirect_browser_to", "use_flow_id"})
			})

			t.Run("case=validation error", func(t *testing.T) {
				t.Cleanup(reset)

				loginFlow = newFlow(t, time.Minute, tc.t)
				flowError = errors.NewInvalidCredentialsError()
				ct = node.PasswordGroup

				res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()
				require.Equal(t, http.StatusBadRequest, res.StatusCode)

				body, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(gjson.GetBytes(body, "ui.messages.0.id").Int()), "%s", body)
				assert.Equal(t, loginFlow.ID.String(), gjson.GetBytes(body, "id").String())
			})

			t.Run("case=generic error", func(t *testing.T) {
				t.Cleanup(reset)

				loginFlow = newFlow(t, time.Minute, tc.t)
				flowError = herodot.ErrInternalServerError.WithReason("system error")
				ct = node.PasswordGroup

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
		expectLoginUI := func(t *testing.T) (*login.Flow, *http.Response) {
			res, err := http.DefaultClient.Get(ts.URL + "/error")
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowLoginUI().String()+"?flow=")

			lf, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(res.Request.URL.Query().Get("flow")))
			require.NoError(t, err)
			return lf, res
		}

		t.Run("case=expired error", func(t *testing.T) {
			t.Cleanup(reset)

			loginFlow = &login.Flow{Type: flow.TypeBrowser}
			flowError = flow.NewFlowExpiredError(anHourAgo)
			ct = node.PasswordGroup

			lf, _ := expectLoginUI(t)
			require.Len(t, lf.UI.Messages, 1)
			assert.Equal(t, int(text.ErrorValidationLoginFlowExpired), int(lf.UI.Messages[0].ID))
		})

		t.Run("case=validation error", func(t *testing.T) {
			t.Cleanup(reset)

			loginFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = errors.NewInvalidCredentialsError()
			ct = node.PasswordGroup

			lf, _ := expectLoginUI(t)
			require.NotEmpty(t, lf.UI.Nodes, x.MustEncodeJSON(t, lf))
			require.Len(t, lf.UI.Messages, 1, x.MustEncodeJSON(t, lf))
			assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(lf.UI.Messages[0].ID), x.MustEncodeJSON(t, lf))
		})

		t.Run("case=generic error", func(t *testing.T) {
			t.Cleanup(reset)

			loginFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = herodot.ErrInternalServerError.WithReason("system error")
			ct = node.PasswordGroup

			sse, _ := expectErrorUI(t)
			assertx.EqualAsJSON(t, flowError, sse)
		})
	})
}
