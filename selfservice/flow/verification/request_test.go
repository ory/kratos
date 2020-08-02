package verification

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
)

func TestNewRequest(t *testing.T) {
	r := NewRequest(
		time.Minute,
		&http.Request{
			URL:  urlx.ParseOrPanic("/source"),
			Host: "kratos.ory.sh",
			TLS:  new(tls.ConnectionState),
		},
		identity.VerifiableAddressTypeEmail,
		urlx.ParseOrPanic("https://kratos.ory.sh/action"),
		func(r *http.Request) string {
			return "anti-csrf"
		},
	)

	assert.NotEmpty(t, r.ID)
	assert.True(t, r.ExpiresAt.After(time.Now()))
	assert.Equal(t, "https://kratos.ory.sh/source", r.RequestURL)
	assert.Equal(t, "https://kratos.ory.sh/action?request="+r.ID.String(), r.Form.Action)
	assert.Equal(t, "anti-csrf", r.Form.Fields[0].Value)
	assert.Equal(t, "to_verify", r.Form.Fields[1].Name)
	assert.Equal(t, identity.VerifiableAddressTypeEmail, r.Via)
	assert.Equal(t, "anti-csrf", r.CSRFToken)

	t.Run("method=Valid", func(t *testing.T) {
		assert.NoError(t, (&Request{ExpiresAt: time.Now().Add(time.Minute)}).Valid())
		assert.Error(t, (&Request{ExpiresAt: time.Now().Add(-time.Minute)}).Valid())
	})
}
