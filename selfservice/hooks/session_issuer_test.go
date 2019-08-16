package hooks_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/identity"
	"github.com/ory/hive/internal"
	. "github.com/ory/hive/selfservice/hooks"
	"github.com/ory/hive/session"
)

func TestSessionIssuer(t *testing.T) {
	_, reg := internal.NewMemoryRegistry(t)
	viper.Set(configuration.ViperKeyURLsSelfPublic, "http://localhost/")
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/stub.schema.json")

	var r http.Request
	h := NewSessionIssuer(reg)

	t.Run("method=sign-in", func(t *testing.T) {
		w := httptest.NewRecorder()
		sid := uuid.New().String()

		i := identity.NewIdentity("")
		i, err := reg.IdentityPool().Create(context.Background(), i)
		require.NoError(t, err)
		require.NoError(t, h.ExecuteLoginPostHook(w, &r, nil, &session.Session{SID: sid, Identity: i}))

		got, err := reg.SessionManager().Get(context.Background(), sid)
		require.NoError(t, err)
		assert.Equal(t, sid, got.SID)
	})

	t.Run("method=sign-up", func(t *testing.T) {
		w := httptest.NewRecorder()
		sid := uuid.New().String()

		i := identity.NewIdentity("")
		i, err := reg.IdentityPool().Create(context.Background(), i)
		require.NoError(t, err)
		require.NoError(t, h.ExecuteRegistrationPostHook(w, &r, nil, &session.Session{SID: sid, Identity: i}))

		got, err := reg.SessionManager().Get(context.Background(), sid)
		require.NoError(t, err)
		assert.Equal(t, sid, got.SID)
	})
}
