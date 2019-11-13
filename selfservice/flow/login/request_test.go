package login_test

import (
	"testing"
	"time"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/login"
)

func TestFakeRequestData(t *testing.T) {
	internal.RegisterFakes()

	var r login.Request
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

	assert.NotEmpty(t, r.RequestHeaders)
	for k, v := range r.RequestHeaders {
		assert.NotEmpty(t, k)
		assert.NotEmpty(t, v)
	}
}

func TestRequest(t *testing.T) {
	r := &login.Request{ID: "request", RequestHeaders: map[string][]string{"foo": {"bar"}}}
	assert.Equal(t, r.ID, r.GetID())
	assert.Empty(t, r.Declassify().RequestHeaders)

	t.Run("case=expired", func(t *testing.T) {
		for _, tc := range []struct {
			r     *login.Request
			valid bool
		}{
			{
				r:     &login.Request{ExpiresAt: time.Now().Add(time.Hour), IssuedAt: time.Now().Add(-time.Minute)},
				valid: true,
			},
			{r: &login.Request{ExpiresAt: time.Now().Add(time.Hour), IssuedAt: time.Now().Add(time.Minute)}},
			{r: &login.Request{ExpiresAt: time.Now().Add(-time.Hour), IssuedAt: time.Now().Add(-time.Minute)}},
		} {
			if tc.valid {
				require.NoError(t, tc.r.Valid())
			} else {
				require.Error(t, tc.r.Valid())
			}
		}
	})
}
