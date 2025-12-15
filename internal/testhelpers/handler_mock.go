// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
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
	identity.ManagementProvider
	session.ManagementProvider
	session.PersistenceProvider
	config.Provider
}

func MockSetSession(t *testing.T, reg mockDeps, conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		i.NID = uuid.Must(uuid.NewV4())
		require.NoError(t, i.SetCredentialsWithConfig(
			identity.CredentialsTypePassword,
			identity.Credentials{
				Type:        identity.CredentialsTypePassword,
				Identifiers: []string{faker.Email()},
			},
			json.RawMessage(`{"hashed_password":"$"}`)))
		require.NoError(t, reg.IdentityManager().Create(context.Background(), i))

		MockSetSessionWithIdentity(t, reg, conf, i)(w, r)
	}
}

func MockSetSessionWithIdentity(t *testing.T, reg mockDeps, _ *config.Config, i *identity.Identity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activeSession, err := NewActiveSession(r, reg, i, time.Now().UTC(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		require.NoError(t, err)
		if aal := r.URL.Query().Get("set_aal"); len(aal) > 0 {
			activeSession.AuthenticatorAssuranceLevel = identity.AuthenticatorAssuranceLevel(aal)
		}
		require.NoError(t, reg.SessionManager().UpsertAndIssueCookie(context.Background(), w, r, activeSession))

		w.WriteHeader(http.StatusOK)
	}
}

type router interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
}

func MockMakeAuthenticatedRequest(t *testing.T, reg mockDeps, conf *config.Config, router router, req *http.Request) ([]byte, *http.Response) {
	return MockMakeAuthenticatedRequestWithClient(t, reg, conf, router, req, NewClientWithCookies(t))
}

func MockMakeAuthenticatedRequestWithClient(t *testing.T, reg mockDeps, conf *config.Config, router router, req *http.Request, client *http.Client) ([]byte, *http.Response) {
	return MockMakeAuthenticatedRequestWithClientAndID(t, reg, conf, router, req, client, nil)
}

func MockMakeAuthenticatedRequestWithClientAndID(t *testing.T, reg mockDeps, conf *config.Config, router router, req *http.Request, client *http.Client, id *identity.Identity) ([]byte, *http.Response) {
	set := "/" + uuid.Must(uuid.NewV4()).String() + "/set"
	if id == nil {
		router.HandleFunc("GET "+set, MockSetSession(t, reg, conf))
	} else {
		router.HandleFunc("GET "+set, MockSetSessionWithIdentity(t, reg, conf, id))
	}

	MockHydrateCookieClient(t, client, "http://"+req.URL.Host+set+"?"+req.URL.Query().Encode())

	res, err := client.Do(req)
	require.NoError(t, errors.WithStack(err))

	body, err := io.ReadAll(res.Body)
	require.NoError(t, errors.WithStack(err))

	require.NoError(t, res.Body.Close())

	return body, res
}

func NewClientWithCookies(t *testing.T) *http.Client {
	cj, err := cookiejar.New(&cookiejar.Options{})
	require.NoError(t, err)
	return &http.Client{Jar: cj}
}

func NewNoRedirectClientWithCookies(t *testing.T) *http.Client {
	cj, err := cookiejar.New(&cookiejar.Options{})
	require.NoError(t, err)
	return &http.Client{
		Jar: cj,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func MockHydrateCookieClient(t *testing.T, c *http.Client, u string) *http.Cookie {
	var sessionCookie *http.Cookie
	res, err := c.Get(u)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()
	body := x.MustReadAll(res.Body)
	assert.EqualValues(t, http.StatusOK, res.StatusCode)

	var found bool
	for _, rc := range res.Cookies() {
		if rc.Name == config.DefaultSessionCookieName {
			found = true
			sessionCookie = rc
		}
	}
	require.True(t, found, "got body: %s\ngot url: %s", body, res.Request.URL.String())
	return sessionCookie
}

func MockSessionCreateHandlerWithIdentity(t *testing.T, reg mockDeps, i *identity.Identity) (http.HandlerFunc, *session.Session) {
	var ct []identity.CredentialsType

	// if identity was not created with any credentials,
	// then assume a 'password' credential type
	if len(i.Credentials) == 0 {
		return MockSessionCreateHandlerWithIdentityAndAMR(t, reg, i, []identity.CredentialsType{"password"})
	}

	// otherwise, mock session with appropriate credential types
	for _, c := range i.Credentials {
		ct = append(ct, c.Type)
	}

	return MockSessionCreateHandlerWithIdentityAndAMR(t, reg, i, ct)
}

func MockSessionCreateHandlerWithIdentityAndAMR(t *testing.T, reg mockDeps, i *identity.Identity, methods []identity.CredentialsType) (http.HandlerFunc, *session.Session) {
	var sess session.Session
	require.NoError(t, faker.FakeData(&sess))
	// require AuthenticatedAt to be time.Now() as we always compare it to the current time
	sess.AuthenticatedAt = time.Now().UTC()
	sess.IssuedAt = time.Now().UTC()
	sess.ExpiresAt = time.Now().UTC().Add(time.Hour * 24)
	sess.Active = true

	for _, m := range methods {
		if m == identity.CredentialsTypeOIDC {
			if c, ok := i.Credentials[m]; ok {
				var target identity.CredentialsOIDC
				if err := json.Unmarshal(c.Config, &target); err == nil {
					for _, t := range target.Providers {
						sess.CompletedLoginForWithProvider(c.Type, identity.AuthenticatorAssuranceLevel1, t.Provider, "")
					}
					continue
				}
			}
		}
		sess.CompletedLoginFor(m, "")
	}

	sess.SetAuthenticatorAssuranceLevel()

	ctx := context.Background()
	if _, err := reg.Config().DefaultIdentityTraitsSchemaURL(ctx); err != nil {
		SetDefaultIdentitySchema(reg.Config(), "file://./stub/fake-session.schema.json")
	}

	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

	inserted, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), i.ID)
	require.NoError(t, err)
	sess.Identity = inserted

	require.NoError(t, reg.SessionPersister().UpsertSession(context.Background(), &sess))
	require.Len(t, inserted.Credentials, len(i.Credentials))

	return func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, reg.SessionManager().IssueCookie(context.Background(), w, r, &sess))
	}, &sess
}

func MockSessionCreateHandler(t *testing.T, reg mockDeps) (http.HandlerFunc, *session.Session) {
	return MockSessionCreateHandlerWithIdentity(t, reg, &identity.Identity{
		ID: x.NewUUID(), State: identity.StateActive, Traits: identity.Traits(`{"baz":"bar","foo":true,"bar":2.5}`),
	})
}
