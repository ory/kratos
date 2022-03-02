package flowhelpers_test

import (
	"context"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flowhelpers"
	"github.com/ory/kratos/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGuessForcedLoginIdentifier(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/login.schema.json")

	i := identity.NewIdentity("")
	i.Credentials[identity.CredentialsTypePassword] = identity.Credentials{
		Type:        identity.CredentialsTypePassword,
		Identifiers: []string{"foobar"},
	}
	require.NoError(t, reg.IdentityManager().Create(context.Background(), i))

	sess, err := session.NewActiveSession(i, conf, time.Now(), identity.CredentialsTypePassword)
	require.NoError(t, err)
	reg.SessionPersister().UpsertSession(context.Background(), sess)

	r := httptest.NewRequest("GET", "/login", nil)
	r.Header.Set("Authorization", "Bearer "+sess.Token)

	var f login.Flow
	f.Refresh = true

	assert.Equal(t, "foobar", flowhelpers.GuessForcedLoginIdentifier(r, reg, &f, identity.CredentialsTypePassword))
}
