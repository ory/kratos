package hook_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/ory/kratos/corpx"

	"github.com/bxcodec/faker/v3"
	"github.com/gobuffalo/httptest"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/session"
	"github.com/ory/x/sqlcon"
)

func init() {
	corpx.RegisterFakes()
}

func TestSessionDestroyer(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)

	conf.MustSet(config.ViperKeyPublicBaseURL, "http://localhost/")
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/stub.schema.json")

	h := hook.NewSessionDestroyer(reg)

	t.Run("method=ExecuteLoginPostHook", func(t *testing.T) {
		var i identity.Identity
		require.NoError(t, faker.FakeData(&i))
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))

		sessions := make([]session.Session, 5)
		for k := range sessions {
			s := sessions[k] // keep this for pointers' sake ;)
			require.NoError(t, faker.FakeData(&s))
			s.IdentityID = uuid.Nil
			s.Identity = &i

			require.NoError(t, reg.SessionPersister().UpsertSession(context.Background(), &s))
		}

		// Should revoke all the sessions.
		require.NoError(t, h.ExecuteLoginPostHook(
			httptest.NewRecorder(),
			new(http.Request),
			nil,
			&session.Session{Identity: &i},
		))

		for k := range sessions {
			_, err := reg.SessionPersister().GetSession(context.Background(), sessions[k].ID)
			assert.EqualError(t, err, sqlcon.ErrNoRows.Error())
		}
	})
}
