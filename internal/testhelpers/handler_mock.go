package testhelpers

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/ory/kratos/internal"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type mockDeps interface {
	identity.PrivilegedPoolProvider
	session.ManagementProvider
	session.PersistenceProvider
	Configuration() *config.Config
}

func MockSetSession(t *testing.T, reg mockDeps, conf *config.Config) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

		require.NoError(t, reg.SessionManager().CreateAndIssueCookie(context.Background(), w, r, session.NewActiveSession(i, conf, time.Now().UTC())))

		w.WriteHeader(http.StatusOK)
	}
}

func MockGetSession(t *testing.T, reg mockDeps) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, err := reg.SessionManager().FetchFromRequest(r.Context(), r)
		if r.URL.Query().Get("has") == "yes" {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func MockMakeAuthenticatedRequest(t *testing.T, reg mockDeps, conf *config.Config, router *httprouter.Router, req *http.Request) ([]byte, *http.Response) {
	set := "/" + uuid.New().String() + "/set"
	router.GET(set, MockSetSession(t, reg, conf))

	client := NewClientWithCookies(t)
	MockHydrateCookieClient(t, client, "http://"+req.URL.Host+set)

	res, err := client.Do(req)
	require.NoError(t, errors.WithStack(err))

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, errors.WithStack(err))

	require.NoError(t, res.Body.Close())

	return body, res
}

func NewClientWithCookies(t *testing.T) *http.Client {
	cj, err := cookiejar.New(&cookiejar.Options{})
	require.NoError(t, err)
	return &http.Client{Jar: cj}
}

func MockHydrateCookieClient(t *testing.T, c *http.Client, u string) {
	res, err := c.Get(u)
	require.NoError(t, err)
	defer res.Body.Close()
	assert.EqualValues(t, http.StatusOK, res.StatusCode)

	var found bool
	for _, c := range res.Cookies() {
		if c.Name == session.DefaultSessionCookieName {
			found = true
		}
	}
	require.True(t, found)
}

func MockSessionCreateHandlerWithIdentity(t *testing.T, reg mockDeps, i *identity.Identity) (httprouter.Handle, *session.Session) {
	var sess session.Session
	require.NoError(t, faker.FakeData(&sess))
	// require AuthenticatedAt to be time.Now() as we always compare it to the current time
	sess.AuthenticatedAt = time.Now().UTC()
	sess.IssuedAt = time.Now().UTC()
	sess.ExpiresAt = time.Now().UTC().Add(time.Hour * 24)
	sess.Active = true

	if reg.Configuration().Source().String(config.ViperKeyDefaultIdentitySchemaURL) == internal.UnsetDefaultIdentitySchema {
		reg.Configuration().MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/fake-session.schema.json")
	}

	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

	inserted, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), i.ID)
	require.NoError(t, err)
	sess.Identity = inserted

	require.NoError(t, reg.SessionPersister().CreateSession(context.Background(), &sess))
	require.Len(t, inserted.Credentials, len(i.Credentials))

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NoError(t, reg.SessionManager().IssueCookie(context.Background(), w, r, &sess))
	}, &sess
}

func MockSessionCreateHandler(t *testing.T, reg mockDeps) (httprouter.Handle, *session.Session) {
	return MockSessionCreateHandlerWithIdentity(t, reg, &identity.Identity{
		ID:     x.NewUUID(),
		Traits: identity.Traits(`{"baz":"bar","foo":true,"bar":2.5}`),
	})
}
