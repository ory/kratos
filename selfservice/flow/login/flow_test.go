package login_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/tidwall/gjson"

	"github.com/ory/x/jsonx"

	"github.com/ory/kratos/identity"

	"github.com/ory/kratos/internal"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
)

func TestFakeFlow(t *testing.T) {
	var r login.Flow
	require.NoError(t, faker.FakeData(&r))

	assert.NotEmpty(t, r.ID)
	assert.NotEmpty(t, r.IssuedAt)
	assert.NotEmpty(t, r.ExpiresAt)
	assert.NotEmpty(t, r.RequestURL)
	assert.NotEmpty(t, r.Active)
	assert.NotNil(t, r.UI)
}

func TestNewFlow(t *testing.T) {
	conf, _ := internal.NewFastRegistryWithMocks(t)

	t.Run("type=aal", func(t *testing.T) {
		r, err := login.NewFlow(conf, 0, "csrf", &http.Request{URL: &url.URL{Path: "/", RawQuery: "aal=aal2&refresh=true"}, Host: "ory.sh"}, flow.TypeBrowser)
		require.NoError(t, err)
		assert.True(t, r.Refresh)
		assert.Equal(t, identity.AuthenticatorAssuranceLevel2, r.RequestedAAL)

		r, err = login.NewFlow(conf, 0, "csrf", &http.Request{URL: &url.URL{Path: "/", RawQuery: "refresh=true"}, Host: "ory.sh"}, flow.TypeBrowser)
		require.NoError(t, err)
		assert.True(t, r.Refresh)
		assert.Equal(t, identity.AuthenticatorAssuranceLevel1, r.RequestedAAL)
	})

	t.Run("type=return_to", func(t *testing.T) {
		_, err := login.NewFlow(conf, 0, "csrf", &http.Request{URL: &url.URL{Path: "/", RawQuery: "return_to=https://not-allowed/foobar"}, Host: "ory.sh"}, flow.TypeBrowser)
		require.Error(t, err)

		_, err = login.NewFlow(conf, 0, "csrf", &http.Request{URL: &url.URL{Path: "/", RawQuery: "return_to=" + urlx.AppendPaths(conf.SelfPublicURL(), "/self-service/login/browser").String()}, Host: "ory.sh"}, flow.TypeBrowser)
		require.NoError(t, err)
	})

	t.Run("type=browser", func(t *testing.T) {
		t.Run("case=regular flow creation without a session", func(t *testing.T) {
			r, err := login.NewFlow(conf, 0, "csrf", &http.Request{
				URL:  urlx.ParseOrPanic("/"),
				Host: "ory.sh", TLS: &tls.ConnectionState{},
			}, flow.TypeBrowser)
			require.NoError(t, err)
			assert.EqualValues(t, r.IssuedAt, r.ExpiresAt)
			assert.Equal(t, flow.TypeBrowser, r.Type)
			assert.False(t, r.Refresh)
			assert.Equal(t, "https://ory.sh/", r.RequestURL)
		})

		t.Run("case=regular flow creation", func(t *testing.T) {
			r, err := login.NewFlow(conf, 0, "csrf", &http.Request{
				URL:  urlx.ParseOrPanic("https://ory.sh/"),
				Host: "ory.sh"}, flow.TypeBrowser)
			require.NoError(t, err)
			assert.Equal(t, "https://ory.sh/", r.RequestURL)
		})
	})

	t.Run("type=api", func(t *testing.T) {
		t.Run("case=flow with refresh", func(t *testing.T) {
			r, err := login.NewFlow(conf, 0, "csrf", &http.Request{
				URL:  urlx.ParseOrPanic("/?refresh=true"),
				Host: "ory.sh"}, flow.TypeAPI)
			require.NoError(t, err)
			assert.Equal(t, r.IssuedAt, r.ExpiresAt)
			assert.Equal(t, flow.TypeAPI, r.Type)
			assert.True(t, r.Refresh)
			assert.Equal(t, "http://ory.sh/?refresh=true", r.RequestURL)
		})

		t.Run("case=flow without refresh", func(t *testing.T) {
			r, err := login.NewFlow(conf, 0, "csrf", &http.Request{
				URL:  urlx.ParseOrPanic("/"),
				Host: "ory.sh"}, flow.TypeAPI)
			require.NoError(t, err)
			assert.Equal(t, r.IssuedAt, r.ExpiresAt)
			assert.Equal(t, flow.TypeAPI, r.Type)
			assert.False(t, r.Refresh)
			assert.Equal(t, "http://ory.sh/", r.RequestURL)
		})
	})
}

func TestFlow(t *testing.T) {
	r := &login.Flow{ID: x.NewUUID()}
	assert.Equal(t, r.ID, r.GetID())

	t.Run("case=expired", func(t *testing.T) {
		for _, tc := range []struct {
			r     *login.Flow
			valid bool
		}{
			{
				r:     &login.Flow{ExpiresAt: time.Now().Add(time.Hour), IssuedAt: time.Now().Add(-time.Minute)},
				valid: true,
			},
			{r: &login.Flow{ExpiresAt: time.Now().Add(-time.Hour), IssuedAt: time.Now().Add(-time.Minute)}},
		} {
			if tc.valid {
				require.NoError(t, tc.r.Valid())
			} else {
				require.Error(t, tc.r.Valid())
			}
		}
	})
}

func TestGetType(t *testing.T) {
	for _, ft := range []flow.Type{
		flow.TypeAPI,
		flow.TypeBrowser,
	} {
		t.Run(fmt.Sprintf("case=%s", ft), func(t *testing.T) {
			r := &login.Flow{Type: ft}
			assert.Equal(t, ft, r.GetType())
		})
	}
}

func TestGetRequestURL(t *testing.T) {
	expectedURL := "http://foo/bar/baz"
	f := &login.Flow{RequestURL: expectedURL}
	assert.Equal(t, expectedURL, f.GetRequestURL())
}

func TestFlowEncodeJSON(t *testing.T) {
	assert.EqualValues(t, "", gjson.Get(jsonx.TestMarshalJSONString(t, &login.Flow{RequestURL: "https://foo.bar?foo=bar"}), "return_to").String())
	assert.EqualValues(t, "/bar", gjson.Get(jsonx.TestMarshalJSONString(t, &login.Flow{RequestURL: "https://foo.bar?return_to=/bar"}), "return_to").String())
	assert.EqualValues(t, "/bar", gjson.Get(jsonx.TestMarshalJSONString(t, login.Flow{RequestURL: "https://foo.bar?return_to=/bar"}), "return_to").String())
}
