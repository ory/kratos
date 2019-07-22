package hooks_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/hive-cloud/hive/driver"
	"github.com/ory/hive-cloud/hive/driver/configuration"
	. "github.com/ory/hive-cloud/hive/selfservice/hooks"
	"github.com/ory/hive-cloud/hive/session"
)

func TestSessionIssuer(t *testing.T) {
	conf := configuration.NewViperProvider(logrus.New())
	reg := new(driver.RegistryMemory).WithConfig(conf)

	viper.Set(configuration.ViperKeyURLsSelfPublic, "http://localhost/")

	var r http.Request
	h := NewSessionIssuer(reg)

	t.Run("method=sign-in", func(t *testing.T) {
		w := httptest.NewRecorder()
		sid := uuid.New().String()
		require.NoError(t, h.ExecuteLoginPostHook(w, &r, nil, &session.Session{SID: sid}))

		got, err := reg.SessionManager().Get(sid)
		require.NoError(t, err)
		assert.Equal(t, sid, got.SID)
	})

	t.Run("method=sign-up", func(t *testing.T) {
		w := httptest.NewRecorder()
		sid := uuid.New().String()
		require.NoError(t, h.ExecuteRegistrationPostHook(w, &r, nil, &session.Session{SID: sid}))

		got, err := reg.SessionManager().Get(sid)
		require.NoError(t, err)
		assert.Equal(t, sid, got.SID)
	})
}
