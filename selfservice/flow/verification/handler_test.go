package verification_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gobuffalo/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/x"
)

func TestGetFlow(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	viper.Set(configuration.ViperKeySelfServiceVerificationEnabled, true)
	viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+verification.StrategyVerificationLinkName,
		map[string]interface{}{"enabled": true})
	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

	public, admin := testhelpers.NewKratosServerWithCSRF(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRedirTS(t, "")

	newVerificationTS := func(t *testing.T, upstream string, c *http.Client) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c == nil {
				c = http.DefaultClient
			}
			_, err := w.Write(x.EasyGetBody(t, c, upstream+verification.RouteGetFlow+"?id="+r.URL.Query().Get("flow")))
			require.NoError(t, err)
		}))
	}

	assertFlowPayload := func(t *testing.T, body []byte, isApi bool) {
		if isApi {
			assert.Equal(t, "api", gjson.GetBytes(body, "type").String(), "%s", body)
			assert.Empty(t, gjson.GetBytes(body, "methods.link.config.fields.#(name==csrf_token).value").String(), "%s", body)
		} else {
			assert.Equal(t, "browser", gjson.GetBytes(body, "type").String(), "%s", body)
			assert.NotEmpty(t, gjson.GetBytes(body, "methods.link.config.fields.#(name==csrf_token).value").String(), "%s", body)
		}

		assert.Equal(t, "link", gjson.GetBytes(body, "methods.link.method").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Empty(t, gjson.GetBytes(body, "headers").Value(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "methods.link.config.action").String(), gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "methods.link.config.action").String(), public.URL, "%s", body)
	}

	assertExpiredPayload := func(t *testing.T, res *http.Response, body []byte) {
		assert.EqualValues(t, http.StatusGone, res.StatusCode)
		assert.Equal(t, public.URL+verification.RouteInitBrowserFlow, gjson.GetBytes(body, "error.details.redirect_to").String(), "%s", body)
	}

	newExpiredFlow := func() *verification.Flow {
		return &verification.Flow{
			ID:         x.NewUUID(),
			ExpiresAt:  time.Now().Add(-time.Minute),
			IssuedAt:   time.Now().Add(-time.Minute * 2),
			RequestURL: public.URL + verification.RouteInitBrowserFlow,
			CSRFToken:  x.FakeCSRFToken,
			Type:       flow.TypeBrowser,
		}
	}

	run := func(t *testing.T, endpoint *httptest.Server) {
		verificationTS := newVerificationTS(t, endpoint.URL, nil)
		defer verificationTS.Close()
		viper.Set(configuration.ViperKeySelfServiceVerificationUI, verificationTS.URL)
		viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword),
			map[string]interface{}{"enabled": true})

		t.Run("case=valid", func(t *testing.T) {
			t.Run("type=browser", func(t *testing.T) {
				assertFlowPayload(t, x.EasyGetBody(t, endpoint.Client(), public.URL+verification.RouteInitBrowserFlow), false)
			})

			t.Run("type=api", func(t *testing.T) {
				res, body := x.EasyGet(t, endpoint.Client(), public.URL+verification.RouteInitAPIFlow)
				assert.Len(t, res.Header.Get("Set-Cookie"), 0)
				assertFlowPayload(t, body, true)
			})
		})

		t.Run("case=expired", func(t *testing.T) {
			lr := newExpiredFlow()
			require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(context.Background(), lr))
			res, body := x.EasyGet(t, admin.Client(), endpoint.URL+verification.RouteGetFlow+"?id="+lr.ID.String())
			assertExpiredPayload(t, res, body)
		})
	}

	t.Run("daemon=admin", func(t *testing.T) {
		run(t, admin)
	})

	t.Run("daemon=public", func(t *testing.T) {
		run(t, public)
	})
}
