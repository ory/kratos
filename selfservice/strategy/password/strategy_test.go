package password_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func newReturnTs(t *testing.T, reg interface {
	session.ManagementProvider
	x.WriterProvider
	config.Provider
}) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := reg.SessionManager().FetchFromRequest(r.Context(), r)
		require.NoError(t, err)
		reg.Writer().Write(w, r, sess)
	}))
	t.Cleanup(ts.Close)
	reg.Config(context.Background()).MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, ts.URL+"/return-ts")
	return ts
}

func TestCountActiveCredentials(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	strategy := password.NewStrategy(reg)

	hash, err := reg.Hasher().Generate(context.Background(), []byte("a password"))
	require.NoError(t, err)

	for k, tc := range []struct {
		in       identity.CredentialsCollection
		expected int
	}{
		{
			in: identity.CredentialsCollection{{
				Type:   strategy.ID(),
				Config: []byte{},
			}},
			expected: 0,
		},
		{
			in: identity.CredentialsCollection{{
				Type:   strategy.ID(),
				Config: []byte(`{"hashed_password": "` + string(hash) + `"}`),
			}},
			expected: 0,
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{""},
				Config:      []byte(`{"hashed_password": "` + string(hash) + `"}`),
			}},
			expected: 0,
		},
		{
			in: identity.CredentialsCollection{{
				Type:        strategy.ID(),
				Identifiers: []string{"foo"},
				Config:      []byte(`{"hashed_password": "` + string(hash) + `"}`),
			}},
			expected: 1,
		},
		{
			in: identity.CredentialsCollection{{
				Type:   strategy.ID(),
				Config: []byte(`{"hashed_password": "asdf"}`),
			}},
			expected: 0,
		},
		{
			in: identity.CredentialsCollection{{
				Type:   strategy.ID(),
				Config: []byte(`{}`),
			}},
			expected: 0,
		},
		{
			in:       identity.CredentialsCollection{{}, {}},
			expected: 0,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			cc := map[identity.CredentialsType]identity.Credentials{}
			for _, c := range tc.in {
				cc[c.Type] = c
			}

			actual, err := strategy.CountActiveCredentials(cc)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

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
