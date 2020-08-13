package registration_test

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/x"
)

func TestFakeFlow(t *testing.T) {
	var r registration.Flow
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
	t.Run("case=0", func(t *testing.T) {
		r := registration.NewFlow(0, "csrf", &http.Request{
			URL:  urlx.ParseOrPanic("/"),
			Host: "ory.sh", TLS: &tls.ConnectionState{},
		}, flow.TypeBrowser)
		assert.Equal(t, r.IssuedAt, r.ExpiresAt)
		assert.Equal(t, flow.TypeBrowser, r.Type)
		assert.Equal(t, "https://ory.sh/", r.RequestURL)
	})

	t.Run("case=1", func(t *testing.T) {
		r := registration.NewFlow(0, "csrf", &http.Request{
			URL:  urlx.ParseOrPanic("/?refresh=true"),
			Host: "ory.sh"}, flow.TypeAPI)
		assert.Equal(t, r.IssuedAt, r.ExpiresAt)
		assert.Equal(t, flow.TypeAPI, r.Type)
		assert.Equal(t, "http://ory.sh/?refresh=true", r.RequestURL)
	})

	t.Run("case=2", func(t *testing.T) {
		r := registration.NewFlow(0, "csrf", &http.Request{
			URL:  urlx.ParseOrPanic("https://ory.sh/"),
			Host: "ory.sh"}, flow.TypeBrowser)
		assert.Equal(t, "http://ory.sh/", r.RequestURL)
	})
}

func TestFlow(t *testing.T) {
	r := &registration.Flow{ID: x.NewUUID()}
	assert.Equal(t, r.ID, r.GetID())

	t.Run("case=expired", func(t *testing.T) {
		for _, tc := range []struct {
			r     *registration.Flow
			valid bool
		}{
			{
				r:     &registration.Flow{ExpiresAt: time.Now().Add(time.Hour), IssuedAt: time.Now().Add(-time.Minute)},
				valid: true,
			},
			{r: &registration.Flow{ExpiresAt: time.Now().Add(-time.Hour), IssuedAt: time.Now().Add(-time.Minute)}},
		} {
			if tc.valid {
				require.NoError(t, tc.r.Valid())
			} else {
				require.Error(t, tc.r.Valid())
			}
		}
	})
}
