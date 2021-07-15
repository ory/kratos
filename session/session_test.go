package session_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/session"
)

func TestSession(t *testing.T) {
	conf, _ := internal.NewFastRegistryWithMocks(t)
	authAt := time.Now()

	i := new(identity.Identity)
	i.State = identity.StateActive
	s, _ := session.NewActiveSession(i, conf, authAt)
	assert.True(t, s.IsActive())

	require.NotEmpty(t, s.Token)
	require.NotEmpty(t, s.LogoutToken)

	i = new(identity.Identity)
	s, err := session.NewActiveSession(i, conf, authAt)
	assert.Nil(t, s)
	assert.Equal(t, session.ErrIdentityDisabled, err)

	assert.False(t, (&session.Session{ExpiresAt: time.Now().Add(time.Hour)}).IsActive())
	assert.False(t, (&session.Session{Active: true}).IsActive())
}
