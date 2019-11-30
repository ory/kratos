package session_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlcon/dockertest"

	"github.com/ory/viper"

	"github.com/ory/herodot"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	. "github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

// nolint: staticcheck
func TestMain(m *testing.M) {
	flag.Parse()
	runner := dockertest.Register()
	runner.Exit(m.Run())
}

func fakeIdentity(t *testing.T, reg Registry) *identity.Identity {
	i := &identity.Identity{
		ID:              x.NewUUID(),
		TraitsSchemaURL: "file://./stub/identity.schema.json",
		Traits:          json.RawMessage(`{}`),
	}

	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/identity.schema.json")

	require.NoError(t, reg.IdentityPool().CreateIdentity(context.Background(), i))
	return i
}

func TestSessionManager(t *testing.T) {
	_, reg := internal.NewMemoryRegistry(t)
	registries := map[string]Registry{
		"memory": reg,
	}

	if !testing.Short() {
		var l sync.Mutex
		dockertest.Parallel([]func(){
			func() {
				db, err := dockertest.ConnectToTestPostgreSQL()
				require.NoError(t, err)

				_, reg := internal.NewRegistrySQL(t, db)

				l.Lock()
				registries["postgres"] = reg
				l.Unlock()
			},
		})
	}

	for name, sm := range registries {
		t.Run(fmt.Sprintf("manager=%s", name), func(t *testing.T) {
			_, err := sm.SessionManager().GetSession(context.Background(), "does-not-exist")
			require.Error(t, err)

			var gave Session
			require.NoError(t, faker.FakeData(&gave))
			gave.Identity = fakeIdentity(t, registries[name])

			require.NoError(t, sm.SessionManager().CreateSession(context.Background(), &gave))

			got, err := sm.SessionManager().GetSession(context.Background(), gave.ID)
			require.NoError(t, err)
			assert.Equal(t, gave.Identity.ID, got.Identity.ID)
			assert.Equal(t, gave.ID, got.ID)
			assert.EqualValues(t, gave.ExpiresAt.Unix(), got.ExpiresAt.Unix())
			assert.Equal(t, gave.AuthenticatedAt.Unix(), got.AuthenticatedAt.Unix())
			assert.Equal(t, gave.IssuedAt.Unix(), got.IssuedAt.Unix())

			require.NoError(t, sm.SessionManager().DeleteSession(context.Background(), gave.ID))
			_, err = sm.SessionManager().GetSession(context.Background(), gave.ID)
			require.Error(t, err)
		})
	}
}

func TestSessionManagerHTTP(t *testing.T) {
	conf, reg := internal.NewMemoryRegistry(t)
	sm := NewManagerMemory(conf, reg)
	h := herodot.NewJSONWriter(nil)

	var s Session
	require.NoError(t, faker.FakeData(&s))
	s.Identity = fakeIdentity(t, reg)

	router := httprouter.New()
	router.GET("/set", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NoError(t, sm.CreateSession(context.Background(), &s))
		require.NoError(t, sm.SaveToRequest(context.Background(), &s, w, r))
	})

	router.GET("/set-direct", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, err := sm.CreateToRequest(context.Background(), s.Identity, w, r)
		require.NoError(t, err)
	})

	router.GET("/clear", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NoError(t, sm.PurgeFromRequest(context.Background(), w, r))
	})

	router.GET("/get", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		s, err := sm.FetchFromRequest(context.Background(), w, r)
		if errorsx.Cause(err) == ErrNoActiveSessionFound {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			assert.NoError(t, err)
			return
		}

		h.Write(w, r, s)
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("case=exists", func(t *testing.T) {
		c := &http.Client{}
		c.Jar, _ = cookiejar.New(nil)

		res, err := c.Get(ts.URL + "/set")
		require.NoError(t, err)
		defer res.Body.Close()
		require.EqualValues(t, http.StatusOK, res.StatusCode)

		res, err = c.Get(ts.URL + "/get")
		require.NoError(t, err)
		defer res.Body.Close()
		require.EqualValues(t, http.StatusOK, res.StatusCode)

		var got Session
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
	})

	t.Run("case=exists-direct", func(t *testing.T) {
		c := &http.Client{}
		c.Jar, _ = cookiejar.New(nil)

		res, err := c.Get(ts.URL + "/set-direct")
		require.NoError(t, err)
		defer res.Body.Close()
		require.EqualValues(t, http.StatusOK, res.StatusCode)

		res, err = c.Get(ts.URL + "/get")
		require.NoError(t, err)
		defer res.Body.Close()
		require.EqualValues(t, http.StatusOK, res.StatusCode)

		var got Session
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
	})

	t.Run("case=exists_clears", func(t *testing.T) {
		c := &http.Client{}
		c.Jar, _ = cookiejar.New(nil)

		res, err := c.Get(ts.URL + "/set")
		require.NoError(t, err)
		defer res.Body.Close()
		require.EqualValues(t, http.StatusOK, res.StatusCode)

		res, err = c.Get(ts.URL + "/clear")
		require.NoError(t, err)
		defer res.Body.Close()
		require.EqualValues(t, http.StatusOK, res.StatusCode)

		res, err = c.Get(ts.URL + "/get")
		require.NoError(t, err)
		defer res.Body.Close()
		assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("case=not-exist", func(t *testing.T) {
		c := &http.Client{}
		c.Jar, _ = cookiejar.New(nil)

		res, err := c.Get(ts.URL + "/get")
		require.NoError(t, err)
		defer res.Body.Close()
		assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
	})
}
