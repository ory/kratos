package password_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func expectStatusCodeBrowserOKOr(isAPI bool, apiExpect int) int {
	return expectStatusCode(isAPI, apiExpect, http.StatusOK)
}

func expectStatusCode(isAPI bool, api, browser int) int {
	if isAPI {
		return api
	}
	return browser
}

func newReturnTs(t *testing.T, reg interface {
	session.ManagementProvider
	x.WriterProvider
}) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := reg.SessionManager().FetchFromRequest(r.Context(), r)
		require.NoError(t, err)
		reg.Writer().Write(w, r, sess)
	}))
	t.Cleanup(ts.Close)
	viper.Set(configuration.ViperKeySelfServiceBrowserDefaultReturnTo, ts.URL+"/return-ts")
	return ts
}

func TestCountActiveCredentials(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	strategy := password.NewStrategy(reg, conf)

	hash, err := reg.Hasher().Generate([]byte("a password"))
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
