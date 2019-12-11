package hook_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func TestSessionDestroyer(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)
	viper.Set(configuration.ViperKeyURLsSelfPublic, "http://localhost/")
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/stub.schema.json")

	var r http.Request
	h := hook.NewSessionIssuer(reg)

	t.Run("method=ExecuteLoginPostHook", func(t *testing.T) {
		w := httptest.NewRecorder()
		sid := x.NewUUID()

		i := identity.NewIdentity("")
		require.NoError(t, reg.IdentityPool().CreateIdentity(context.Background(), i))
		err := reg.SessionPersister().CreateSession(context.Background(), &session.Session{ID: sid, Identity: i}))
		require.NoError(t, err)
		err = reg.SessionPersister().CreateSession(context.Background(), &session.Session{ID: sid, Identity: i}))
		require.NoError(t, err)		
		require.NoError(t, h.ExecuteLoginPostHook(w, &r, nil, &session.Session{ID: sid, Identity: i}))

		_, err := reg.SessionPersister().GetSession(context.Background(), sid)
		require.Error(t, err)
	})

}
