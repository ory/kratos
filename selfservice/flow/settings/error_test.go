package settings_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/ui/node"

	"github.com/bxcodec/faker/v3"
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
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func TestHandleError(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

	public, admin := testhelpers.NewKratosServer(t, reg)

	router := httprouter.New()
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	testhelpers.NewErrorTestServer(t, reg)
	testhelpers.NewLoginUIFlowEchoServer(t, reg)

	h := reg.SettingsFlowErrorHandler()
	sdk := testhelpers.NewSDKClient(admin)

	var settingsFlow *settings.Flow
	var flowError error
	var flowMethod node.Group
	var id identity.Identity
	require.NoError(t, faker.FakeData(&id))
	id.SchemaID = "default"
	id.State = identity.StateActive

	router.GET("/error", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h.WriteFlowError(w, r, flowMethod, settingsFlow, &id, flowError)
	})

	reset := func() {
		settingsFlow = nil
		flowError = nil
		flowMethod = node.DefaultGroup
	}

	newFlow := func(t *testing.T, ttl time.Duration, ft flow.Type) *settings.Flow {
		req := &http.Request{URL: urlx.ParseOrPanic("/")}
		f := settings.NewFlow(conf, ttl, req, &id, ft)
		for _, s := range reg.SettingsStrategies(context.Background()) {
			require.NoError(t, s.PopulateSettingsMethod(req, &id, f))
		}

		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(context.Background(), f))
		return f
	}

	expectErrorUI := func(t *testing.T) ([]map[string]interface{}, *http.Response) {
		res, err := ts.Client().Get(ts.URL + "/error")
		require.NoError(t, err)
		defer res.Body.Close()
		require.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowErrorURL().String()+"?error=")

		sse, _, err := sdk.PublicApi.GetSelfServiceError(context.Background()).
			Error_(res.Request.URL.Query().Get("error")).Execute()
		require.NoError(t, err)

		return sse.Errors, nil
	}

	expiredAnHourAgo := time.Now().Add(-time.Hour)

	t.Run("case=error with nil flow defaults to error ui redirect", func(t *testing.T) {
		t.Cleanup(reset)

		flowError = herodot.ErrInternalServerError.WithReason("system error")
		flowMethod = settings.StrategyProfile

		sse, _ := expectErrorUI(t)
		assertx.EqualAsJSON(t, []interface{}{flowError}, sse)
	})

	t.Run("case=error with nil flow detects application/json", func(t *testing.T) {
		t.Cleanup(reset)

		flowError = herodot.ErrInternalServerError.WithReason("system error")
		flowMethod = settings.StrategyProfile

		res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
		require.NoError(t, err)
		defer res.Body.Close()
		assert.Contains(t, res.Header.Get("Content-Type"), "application/json")
		assert.NotContains(t, res.Request.URL.String(), conf.SelfServiceFlowErrorURL().String()+"?error=")

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

				// This needs an authenticated client in order to call the RouteGetFlow endpoint
				s, err := session.NewActiveSession(&id, testhelpers.NewSessionLifespanProvider(time.Hour), time.Now())
				require.NoError(t, err)
				c := testhelpers.NewHTTPClientWithSessionToken(t, reg, s)

				settingsFlow = newFlow(t, time.Minute, tc.t)
				flowError = settings.NewFlowExpiredError(expiredAnHourAgo)
				flowMethod = settings.StrategyProfile

				res, err := c.Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()
				require.Contains(t, res.Request.URL.String(), public.URL+settings.RouteGetFlow)

				body, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, res.StatusCode, "%+v\n\t%s", res.Request, body)

				assert.Equal(t, int(text.ErrorValidationSettingsFlowExpired), int(gjson.GetBytes(body, "ui.messages.0.id").Int()), "%s", body)
				assert.NotEqual(t, settingsFlow.ID.String(), gjson.GetBytes(body, "id").String())
			})

			t.Run("case=validation error", func(t *testing.T) {
				t.Cleanup(reset)

				settingsFlow = newFlow(t, time.Minute, tc.t)
				flowError = schema.NewInvalidCredentialsError()
				flowMethod = settings.StrategyProfile

				res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()
				require.Equal(t, http.StatusBadRequest, res.StatusCode)

				body, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(gjson.GetBytes(body, "ui.messages.0.id").Int()), "%s", body)
				assert.Equal(t, settingsFlow.ID.String(), gjson.GetBytes(body, "id").String())
			})

			t.Run("case=generic error", func(t *testing.T) {
				t.Cleanup(reset)

				settingsFlow = newFlow(t, time.Minute, tc.t)
				flowError = herodot.ErrInternalServerError.WithReason("system error")
				flowMethod = settings.StrategyProfile

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
		expectSettingsUI := func(t *testing.T) (*settings.Flow, *http.Response) {
			res, err := ts.Client().Get(ts.URL + "/error")
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowSettingsUI().String()+"?flow=")

			sf, err := reg.SettingsFlowPersister().GetSettingsFlow(context.Background(), uuid.FromStringOrNil(res.Request.URL.Query().Get("flow")))
			require.NoError(t, err)
			return sf, res
		}

		t.Run("case=expired error", func(t *testing.T) {
			t.Cleanup(reset)

			settingsFlow = &settings.Flow{Type: flow.TypeBrowser}
			flowError = settings.NewFlowExpiredError(expiredAnHourAgo)
			flowMethod = settings.StrategyProfile

			lf, _ := expectSettingsUI(t)
			require.Len(t, lf.UI.Messages, 1)
			assert.Equal(t, int(text.ErrorValidationSettingsFlowExpired), int(lf.UI.Messages[0].ID))
		})

		t.Run("case=session old error", func(t *testing.T) {
			t.Cleanup(reset)

			settingsFlow = &settings.Flow{Type: flow.TypeBrowser}
			flowError = settings.NewFlowNeedsReAuth()
			flowMethod = settings.StrategyProfile

			res, err := ts.Client().Get(ts.URL + "/error")
			require.NoError(t, err)
			defer res.Body.Close()
			require.Contains(t, res.Request.URL.String(), conf.Source().String(config.ViperKeySelfServiceLoginUI))
		})

		t.Run("case=validation error", func(t *testing.T) {
			t.Cleanup(reset)

			settingsFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = schema.NewInvalidCredentialsError()
			flowMethod = settings.StrategyProfile

			lf, _ := expectSettingsUI(t)
			require.NotEmpty(t, lf.UI, x.MustEncodeJSON(t, lf))
			require.Len(t, lf.UI.Messages, 1, x.MustEncodeJSON(t, lf))
			assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(lf.UI.Messages[0].ID), x.MustEncodeJSON(t, lf))
		})

		t.Run("case=inaccessible public URL", func(t *testing.T) {
			t.Cleanup(reset)

			// Since WriteFlowError is invoked directly by the /error handler above,
			// manipulate the schema's URL directly to a bad URL.
			id.SchemaURL = "http://some.random.url"

			settingsFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = schema.NewInvalidCredentialsError()
			flowMethod = settings.StrategyProfile

			lf, _ := expectSettingsUI(t)
			require.NotEmpty(t, lf.UI, x.MustEncodeJSON(t, lf))
			require.Len(t, lf.UI.Messages, 1, x.MustEncodeJSON(t, lf))
			assert.Equal(t, int(text.ErrorValidationInvalidCredentials), int(lf.UI.Messages[0].ID), x.MustEncodeJSON(t, lf))
		})

		t.Run("case=generic error", func(t *testing.T) {
			t.Cleanup(reset)

			settingsFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			flowError = herodot.ErrInternalServerError.WithReason("system error")
			flowMethod = settings.StrategyProfile

			sse, _ := expectErrorUI(t)
			assertx.EqualAsJSON(t, []interface{}{flowError}, sse)
		})
	})
}
