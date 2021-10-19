package password_test

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
)

func TestDisabledEndpoint(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypePassword.String(), false)
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://stub/sort.schema.json")

	publicTS, _ := testhelpers.NewKratosServer(t, reg)

	c := testhelpers.NewClientWithCookies(t)
	t.Run("case=should not login when password method is disabled", func(t *testing.T) {
		f := testhelpers.InitializeLoginFlowViaAPI(t, c, publicTS, false)

		res, err := c.PostForm(f.Ui.Action, url.Values{"method": {"password"}, "password_identifier": []string{"identifier"}, "password": []string{"password"}})
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)

		defer res.Body.Close()
		b, err := ioutil.ReadAll(res.Body)
		assert.Contains(t, string(b), "This endpoint was disabled by system administrator", "%s", b)
	})

	t.Run("case=should not registration when password method is disabled", func(t *testing.T) {
		f := testhelpers.InitializeRegistrationFlowViaAPI(t, c, publicTS)

		res, err := c.PostForm(f.Ui.Action, url.Values{"method": {"password"}, "password_identifier": []string{"identifier"}, "password": []string{"password"}})
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)

		defer res.Body.Close()
		b, err := ioutil.ReadAll(res.Body)
		assert.Contains(t, string(b), "This endpoint was disabled by system administrator", "%s", b)
	})

	t.Run("case=should not settings when password method is disabled", func(t *testing.T) {
		require.NoError(t, conf.Set(config.ViperKeyDefaultIdentitySchemaURL, "file://stub/login.schema.json"))
		c := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg)

		t.Run("method=GET", func(t *testing.T) {
			t.Skip("GET is currently not supported for this endpoint.")
		})

		t.Run("method=POST", func(t *testing.T) {
			f := testhelpers.InitializeSettingsFlowViaAPI(t, c, publicTS)
			res, err := c.PostForm(f.Ui.Action, url.Values{
				"method":   {"password"},
				"password": {"bar"},
			})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, res.StatusCode)

			defer res.Body.Close()
			b, err := ioutil.ReadAll(res.Body)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator", "%s", b)
		})
	})
}
