package testhelpers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
)

func CreateSession(t *testing.T, reg driver.Registry) *session.Session {
	ctx := context.Background()
	i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(ctx, i))
	sess, err := session.NewActiveSession(i, reg.Config(ctx), time.Now().UTC())
	require.NoError(t, err)
	require.NoError(t, reg.SessionPersister().CreateSession(ctx, sess))
	return sess
}
