package hook_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/session"
)

func TestSessionIssuer(t *testing.T) {
	_, reg := internal.NewMemoryRegistry(t)
	viper.Set(configuration.ViperKeyURLsSelfPublic, "http://localhost/")
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/stub.schema.json")

	var r http.Request
	h := hook.NewSessionIssuer(reg)

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
