package hook_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/randx"
)

func TestSessionIssuer(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeyPublicBaseURL, "http://localhost/")
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/stub.schema.json")

	var r http.Request
	h := hook.NewSessionIssuer(reg)

	t.Run("method=sign-up", func(t *testing.T) {
		t.Run("flow=browser", func(t *testing.T) {
			w := httptest.NewRecorder()
			sid := x.NewUUID()

			i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
			require.NoError(t, h.ExecutePostRegistrationPostPersistHook(w, &r,
				&registration.Flow{Type: flow.TypeBrowser}, &session.Session{ID: sid, Identity: i, Token: randx.MustString(12, randx.AlphaLowerNum)}))

			got, err := reg.SessionPersister().GetSession(context.Background(), sid)
			require.NoError(t, err)
			assert.Equal(t, sid, got.ID)
			assert.True(t, got.AuthenticatedAt.After(time.Now().Add(-time.Minute)))

			assert.Contains(t, w.Header().Get("Set-Cookie"), session.DefaultSessionCookieName)
		})

		t.Run("flow=api", func(t *testing.T) {
			w := httptest.NewRecorder()

			i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			s := &session.Session{ID: x.NewUUID(), Identity: i, Token: randx.MustString(12, randx.AlphaLowerNum)}
			f := &registration.Flow{Type: flow.TypeAPI}

			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
			err := h.ExecutePostRegistrationPostPersistHook(w, &r, f, s)
			require.True(t, errors.Is(err, registration.ErrHookAbortFlow), "%+v", err)

			got, err := reg.SessionPersister().GetSession(context.Background(), s.ID)
			require.NoError(t, err)
			assert.Equal(t, s.ID.String(), got.ID.String())
			assert.True(t, got.AuthenticatedAt.After(time.Now().Add(-time.Minute)))

			assert.Empty(t, w.Header().Get("Set-Cookie"))
			body := w.Body.Bytes()
			assert.Equal(t, i.ID.String(), gjson.GetBytes(body, "identity.id").String())
			assert.Equal(t, s.ID.String(), gjson.GetBytes(body, "session.id").String())
			assert.Equal(t, got.Token, gjson.GetBytes(body, "session_token").String())
		})
	})
}
