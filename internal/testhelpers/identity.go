package testhelpers

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
)

func CreateSession(t *testing.T, reg driver.Registry) *session.Session {
	req := httptest.NewRequest("GET", "/sessions/whoami", nil)
	i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(req.Context(), i))
	sess, err := session.NewActiveSession(req, i, reg.Config(), time.Now().UTC(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
	require.NoError(t, err)
	require.NoError(t, reg.SessionPersister().UpsertSession(req.Context(), sess))
	return sess
}
