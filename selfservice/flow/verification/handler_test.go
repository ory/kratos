package verification_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/gobuffalo/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/x"
)

func TestGetFlow(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceVerificationEnabled, true)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+verification.StrategyVerificationLinkName,
		map[string]interface{}{"enabled": true})
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

	public, _ := testhelpers.NewKratosServerWithCSRF(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRedirTS(t, "", conf)

	setupVerificationUI := func(t *testing.T, c *http.Client) *httptest.Server {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(x.EasyGetBody(t, c, public.URL+verification.RouteGetFlow+"?id="+r.URL.Query().Get("flow")))
			require.NoError(t, err)
		}))
		t.Cleanup(ts.Close)
		conf.MustSet(config.ViperKeySelfServiceVerificationUI, ts.URL)
		return ts
	}

	assertFlowPayload := func(t *testing.T, body []byte, isApi bool) {
		if isApi {
			assert.Equal(t, "api", gjson.GetBytes(body, "type").String(), "%s", body)
			assert.Empty(t, gjson.GetBytes(body, "ui.fields.#(attributes.name==csrf_token).attributes.value").String(), "%s", body)
		} else {
			assert.Equal(t, "browser", gjson.GetBytes(body, "type").String(), "%s", body)
			assert.NotEmpty(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String(), "%s", body)
		}

		assert.NotEmpty(t, gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Empty(t, gjson.GetBytes(body, "headers").Value(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), public.URL, "%s", body)
	}

	t.Run("case=valid", func(t *testing.T) {
		t.Run("type=browser", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			_ = setupVerificationUI(t, client)
			res, body := x.EasyGet(t, client, public.URL+verification.RouteInitBrowserFlow)
			require.NotEqualValues(t, res.Request.URL.String(), public.URL+verification.RouteInitBrowserFlow)
			assertFlowPayload(t, body, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			_ = setupVerificationUI(t, client)
			res, body := x.EasyGetJSON(t, client, public.URL+verification.RouteInitBrowserFlow)
			require.EqualValues(t, res.Request.URL.String(), public.URL+verification.RouteInitBrowserFlow)
			assertFlowPayload(t, body, false)
		})

		t.Run("type=api", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			_ = setupVerificationUI(t, client)
			res, body := x.EasyGet(t, client, public.URL+verification.RouteInitAPIFlow)
			assert.Len(t, res.Header.Get("Set-Cookie"), 0)
			assertFlowPayload(t, body, true)
		})
	})

	t.Run("case=csrf cookie missing", func(t *testing.T) {
		client := http.DefaultClient
		_ = setupVerificationUI(t, client)
		body := x.EasyGetBody(t, client, public.URL+verification.RouteInitBrowserFlow)

		assert.EqualValues(t, x.ErrInvalidCSRFToken.ReasonField, gjson.GetBytes(body, "error.reason").String(), "%s", body)
	})

	t.Run("case=expired", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		_ = setupVerificationUI(t, client)
		body := x.EasyGetBody(t, client, public.URL+verification.RouteInitBrowserFlow)

		// Expire the flow
		f, err := reg.VerificationFlowPersister().GetVerificationFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
		require.NoError(t, err)
		f.ExpiresAt = time.Now().Add(-time.Second)
		require.NoError(t, reg.VerificationFlowPersister().UpdateVerificationFlow(context.Background(), f))

		res, body := x.EasyGet(t, client, public.URL+verification.RouteGetFlow+"?id="+f.ID.String())
		assert.EqualValues(t, http.StatusGone, res.StatusCode)
		assert.Equal(t, public.URL+verification.RouteInitBrowserFlow, gjson.GetBytes(body, "error.details.redirect_to").String(), "%s", body)
	})
}
