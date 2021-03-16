package settings_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
	"time"

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
	assert.NotEmpty(t, r.Methods)
	for _, m := range r.Methods {
		assert.NotEmpty(t, m.Method)
		assert.NotEmpty(t, m.Config)
	}
}

func TestNewFlow(t *testing.T) {
	id := &identity.Identity{ID: x.NewUUID()}
	t.Run("case=0", func(t *testing.T) {
		r := settings.NewFlow(0, &http.Request{URL: urlx.ParseOrPanic("/"),
			Host: "ory.sh", TLS: &tls.ConnectionState{}}, id, flow.TypeBrowser)
		assert.Equal(t, r.IssuedAt, r.ExpiresAt)
		assert.Equal(t, flow.TypeBrowser, r.Type)
		assert.Equal(t, "https://ory.sh/", r.RequestURL)
	})

	t.Run("case=1", func(t *testing.T) {
		r := settings.NewFlow(0, &http.Request{
			URL:  urlx.ParseOrPanic("/?refresh=true"),
			Host: "ory.sh"}, id, flow.TypeAPI)
		assert.Equal(t, r.IssuedAt, r.ExpiresAt)
		assert.Equal(t, flow.TypeAPI, r.Type)
		assert.Equal(t, "http://ory.sh/?refresh=true", r.RequestURL)
	})

	t.Run("case=2", func(t *testing.T) {
		r := settings.NewFlow(0, &http.Request{
			URL:  urlx.ParseOrPanic("https://ory.sh/"),
			Host: "ory.sh"}, id, flow.TypeBrowser)
		assert.Equal(t, "https://ory.sh/", r.RequestURL)
	})
}

func TestFlow(t *testing.T) {
	alice := x.NewUUID()
	malice := x.NewUUID()
	for k, tc := range []struct {
		r         *settings.Flow
		s         *session.Session
		expectErr bool
	}{
		{
			r: settings.NewFlow(
				time.Hour,
				&http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"},
				&identity.Identity{ID: alice},
				flow.TypeBrowser,
			),
			s: &session.Session{Identity: &identity.Identity{ID: alice}},
		},
		{
			r: settings.NewFlow(
				time.Hour,
				&http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"},
				&identity.Identity{ID: alice},
				flow.TypeBrowser,
			),
			s:         &session.Session{Identity: &identity.Identity{ID: malice}},
			expectErr: true,
		},
		{
			r: settings.NewFlow(
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
