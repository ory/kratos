package hook_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/gobuffalo/httptest"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/sqlcon"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/session"
)

func init() {
	internal.RegisterFakes()
}

func TestSessionDestroyer(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)

	viper.Set(configuration.ViperKeyPublicBaseURL, "http://localhost/")
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/stub.schema.json")

	h := hook.NewSessionIssuer(reg)

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

			require.NoError(t, reg.SessionPersister().CreateSession(context.Background(), &s))
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
