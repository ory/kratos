package settings_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/pkg/errors"

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

	_, admin := testhelpers.NewKratosServer(t, reg)

	router := httprouter.New()
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	errorTS := testhelpers.NewErrorTestServer(t, reg)
	loginTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)

	h := reg.SettingsFlowErrorHandler()
	sdk := testhelpers.NewSDKClient(admin)

	var settingsFlow *settings.Flow
	var flowError error
	var flowMethod node.Group
	var id identity.Identity
	require.NoError(t, faker.FakeData(&id))
	id.SchemaID = "default"
	id.State = identity.StateActive
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &id))

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
		f, err := settings.NewFlow(conf, ttl, req, &id, ft)
		require.NoError(t, err)

		for _, s := range reg.SettingsStrategies(context.Background()) {
			require.NoError(t, s.PopulateSettingsMethod(req, &id, f))
		}

		require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(context.Background(), f))
		return f
	}

	expectErrorUI := func(t *testing.T) (map[string]interface{}, *http.Response) {
		res, err := ts.Client().Get(ts.URL + "/error")
		require.NoError(t, err)
		defer res.Body.Close()
		require.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowErrorURL().String()+"?id=")

		sse, _, err := sdk.V0alpha2Api.GetSelfServiceError(context.Background()).
			Id(res.Request.URL.Query().Get("id")).Execute()
		require.NoError(t, err)

		return sse.Error, nil
	}

	expiredAnHourAgo := time.Now().Add(-time.Hour)

	t.Run("case=error with nil flow defaults to error ui redirect", func(t *testing.T) {
		t.Cleanup(reset)

		flowError = herodot.ErrInternalServerError.WithReason("system error")
		flowMethod = settings.StrategyProfile

		sse, _ := expectErrorUI(t)
		assertx.EqualAsJSON(t, flowError, sse)
	})

	t.Run("case=error with nil flow detects application/json", func(t *testing.T) {
		t.Cleanup(reset)

		flowError = herodot.ErrInternalServerError.WithReason("system error")
		flowMethod = settings.StrategyProfile

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

				// This needs an authenticated client in order to call the RouteGetFlow endpoint
				s, err := session.NewActiveSession(&id, testhelpers.NewSessionLifespanProvider(time.Hour), time.Now(), identity.CredentialsTypePassword)
				require.NoError(t, err)
				c := testhelpers.NewHTTPClientWithSessionToken(t, reg, s)

				settingsFlow = newFlow(t, time.Minute, tc.t)
				flowError = flow.NewFlowExpiredError(expiredAnHourAgo)
				flowMethod = settings.StrategyProfile

				res, err := c.Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()
				require.Contains(t, res.Request.URL.String(), ts.URL+"/error")

				body, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				require.Equal(t, http.StatusGone, res.StatusCode, "%+v\n\t%s", res.Request, body)

				assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", gjson.GetBytes(body, "use_flow_id").String())
				assertx.EqualAsJSONExcept(t, flow.NewFlowExpiredError(expiredAnHourAgo), json.RawMessage(body), []string{"since", "redirect_browser_to", "use_flow_id"})
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

			t.Run("case=return to UI error", func(t *testing.T) {
				t.Cleanup(reset)

				settingsFlow = newFlow(t, time.Minute, flow.TypeBrowser)
				settingsFlow.IdentityID = id.ID
				flowError = flow.ErrStrategyAsksToReturnToUI
				flowMethod = settings.StrategyProfile

				res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()
				require.Equal(t, http.StatusOK, res.StatusCode)

				body, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, settingsFlow.ID.String(), gjson.GetBytes(body, "id").String())
			})

			t.Run("case=no active session", func(t *testing.T) {
				t.Cleanup(reset)

				settingsFlow = newFlow(t, time.Minute, flow.TypeBrowser)
				settingsFlow.IdentityID = id.ID
				flowError = errors.WithStack(session.NewErrNoActiveSessionFound())
				flowMethod = settings.StrategyProfile

				res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()
				require.Equal(t, http.StatusUnauthorized, res.StatusCode)

				body, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, session.NewErrNoActiveSessionFound().Reason(), gjson.GetBytes(body, "error.reason").String(), "%s", body)
			})

			t.Run("case=aal too low", func(t *testing.T) {
				t.Cleanup(reset)

				settingsFlow = newFlow(t, time.Minute, flow.TypeBrowser)
				settingsFlow.IdentityID = id.ID
				flowError = errors.WithStack(session.NewErrAALNotSatisfied("a"))
				flowMethod = settings.StrategyProfile

				res, err := ts.Client().Do(testhelpers.NewHTTPGetJSONRequest(t, ts.URL+"/error"))
				require.NoError(t, err)
				defer res.Body.Close()
				require.Equal(t, http.StatusForbidden, res.StatusCode)

				body, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				assertx.EqualAsJSON(t, session.NewErrAALNotSatisfied("a"), json.RawMessage(body))
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
			flowError = flow.NewFlowExpiredError(expiredAnHourAgo)
			flowMethod = settings.StrategyProfile

			lf, _ := expectSettingsUI(t)
			require.Len(t, lf.UI.Messages, 1)
			assert.Equal(t, int(text.ErrorValidationSettingsFlowExpired), int(lf.UI.Messages[0].ID))
		})

		t.Run("case=return to ui error", func(t *testing.T) {
			t.Cleanup(reset)

			settingsFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			settingsFlow.IdentityID = id.ID
			flowError = flow.ErrStrategyAsksToReturnToUI
			flowMethod = settings.StrategyProfile

			lf, _ := expectSettingsUI(t)
			assert.EqualValues(t, settingsFlow.ID, lf.ID)
		})

		t.Run("case=no active session error", func(t *testing.T) {
			t.Cleanup(reset)

			settingsFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			settingsFlow.IdentityID = id.ID
			flowError = errors.WithStack(session.NewErrNoActiveSessionFound())
			flowMethod = settings.StrategyProfile

			res, err := ts.Client().Get(ts.URL + "/error")
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())
			assert.Contains(t, res.Request.URL.String(), loginTS.URL)

			lf, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(res.Request.URL.Query().Get("flow")))
			require.NoError(t, err)
			assert.Equal(t, identity.AuthenticatorAssuranceLevel1, lf.RequestedAAL)
		})

		t.Run("case=aal too low", func(t *testing.T) {
			t.Cleanup(reset)

			settingsFlow = newFlow(t, time.Minute, flow.TypeBrowser)
			settingsFlow.IdentityID = id.ID
			flowError = errors.WithStack(session.NewErrAALNotSatisfied(""))
			flowMethod = settings.StrategyProfile

			res, err := ts.Client().Get(ts.URL + "/error")
			require.NoError(t, err)
			assert.Contains(t, res.Request.URL.String(), errorTS.URL)
			body := x.MustReadAll(res.Body)
			require.NoError(t, res.Body.Close())

			// We end up at the error endpoint with an aal2 error message because ts.client has no session.
			assert.Equal(t, "You can not requested a higher AAL (AAL2/AAL3) without an active session.", gjson.GetBytes(body, "reason").String(), "%s", body)
		})

		t.Run("case=session old error", func(t *testing.T) {
			conf.MustSet(config.ViperKeyURLsWhitelistedReturnToDomains, []string{urlx.AppendPaths(conf.SelfPublicURL(), "/error").String()})
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
			assertx.EqualAsJSON(t, flowError, sse)
		})
	})
}
