package settings_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ory/kratos/selfservice/flow/registration"

	"github.com/tidwall/gjson"

	"github.com/ory/x/jsonx"

	"github.com/ory/kratos/internal"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func TestFakeFlow(t *testing.T) {
	var r settings.Flow
	require.NoError(t, faker.FakeData(&r))

	assert.NotEmpty(t, r.ID)
	assert.NotEmpty(t, r.IssuedAt)
	assert.NotEmpty(t, r.ExpiresAt)
	assert.NotEmpty(t, r.RequestURL)
	assert.NotEmpty(t, r.Active)
}

func TestNewFlow(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults(t)

	id := &identity.Identity{ID: x.NewUUID()}
	t.Run("case=0", func(t *testing.T) {
		r, err := settings.NewFlow(conf, 0, &http.Request{URL: urlx.ParseOrPanic("/"),
			Host: "ory.sh", TLS: &tls.ConnectionState{}}, id, flow.TypeBrowser)
		require.NoError(t, err)
		assert.Equal(t, r.IssuedAt, r.ExpiresAt)
		assert.Equal(t, flow.TypeBrowser, r.Type)
		assert.Equal(t, "https://ory.sh/", r.RequestURL)
	})

	t.Run("type=return_to", func(t *testing.T) {
		_, err := registration.NewFlow(conf, 0, "csrf", &http.Request{URL: &url.URL{Path: "/", RawQuery: "return_to=https://not-allowed/foobar"}, Host: "ory.sh"}, flow.TypeBrowser)
		require.Error(t, err)

		_, err = registration.NewFlow(conf, 0, "csrf", &http.Request{URL: &url.URL{Path: "/", RawQuery: "return_to=" + urlx.AppendPaths(conf.SelfPublicURL(), "/self-service/login/browser").String()}, Host: "ory.sh"}, flow.TypeBrowser)
		require.NoError(t, err)
	})

	t.Run("case=1", func(t *testing.T) {
		r, err := settings.NewFlow(conf, 0, &http.Request{
			URL:  urlx.ParseOrPanic("/?refresh=true"),
			Host: "ory.sh"}, id, flow.TypeAPI)
		require.NoError(t, err)
		assert.Equal(t, r.IssuedAt, r.ExpiresAt)
		assert.Equal(t, flow.TypeAPI, r.Type)
		assert.Equal(t, "http://ory.sh/?refresh=true", r.RequestURL)
	})

	t.Run("case=2", func(t *testing.T) {
		r, err := settings.NewFlow(conf, 0, &http.Request{
			URL:  urlx.ParseOrPanic("https://ory.sh/"),
			Host: "ory.sh"}, id, flow.TypeBrowser)
		require.NoError(t, err)
		assert.Equal(t, "https://ory.sh/", r.RequestURL)
	})
}

func TestFlow(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults(t)

	alice := x.NewUUID()
	malice := x.NewUUID()
	for k, tc := range []struct {
		r         *settings.Flow
		s         *session.Session
		expectErr bool
	}{
		{
			r: settings.MustNewFlow(
				conf,
				time.Hour,
				&http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"},
				&identity.Identity{ID: alice},
				flow.TypeBrowser,
			),
			s: &session.Session{Identity: &identity.Identity{ID: alice}},
		},
		{
			r: settings.MustNewFlow(
				conf,
				time.Hour,
				&http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"},
				&identity.Identity{ID: alice},
				flow.TypeBrowser,
			),
			s:         &session.Session{Identity: &identity.Identity{ID: malice}},
			expectErr: true,
		},
		{
			r: settings.MustNewFlow(
				conf,
				-time.Hour,
				&http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"},
				&identity.Identity{ID: alice},
				flow.TypeBrowser,
			),
			s:         &session.Session{Identity: &identity.Identity{ID: alice}},
			expectErr: true,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			err := tc.r.Valid(tc.s)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGetType(t *testing.T) {
	for _, ft := range []flow.Type{
		flow.TypeAPI,
		flow.TypeBrowser,
	} {
		t.Run(fmt.Sprintf("case=%s", ft), func(t *testing.T) {
			r := &settings.Flow{Type: ft}
			assert.Equal(t, ft, r.GetType())
		})
	}
}

func TestGetRequestURL(t *testing.T) {
	expectedURL := "http://foo/bar/baz"
	f := &settings.Flow{RequestURL: expectedURL}
	assert.Equal(t, expectedURL, f.GetRequestURL())
}

func TestEnsureInternalContext(t *testing.T) {
	f := new(settings.Flow)
	f.EnsureInternalContext()
	assert.Equal(t, "{}", string(f.InternalContext))
}

func TestFlowEncodeJSON(t *testing.T) {
	assert.EqualValues(t, "", gjson.Get(jsonx.TestMarshalJSONString(t, &settings.Flow{RequestURL: "https://foo.bar?foo=bar"}), "return_to").String())
	assert.EqualValues(t, "/bar", gjson.Get(jsonx.TestMarshalJSONString(t, &settings.Flow{RequestURL: "https://foo.bar?return_to=/bar"}), "return_to").String())
	assert.EqualValues(t, "/bar", gjson.Get(jsonx.TestMarshalJSONString(t, settings.Flow{RequestURL: "https://foo.bar?return_to=/bar"}), "return_to").String())
}
