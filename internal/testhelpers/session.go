// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ory/nosurf"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/gobuffalo/httptest"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type SessionLifespanProvider struct {
	e time.Duration
}

func (p *SessionLifespanProvider) SessionLifespan(ctx context.Context) time.Duration {
	return p.e
}

func NewSessionLifespanProvider(expiresIn time.Duration) *SessionLifespanProvider {
	return &SessionLifespanProvider{e: expiresIn}
}

func NewSessionClient(t *testing.T, u string) *http.Client {
	c := NewClientWithCookies(t)
	MockHydrateCookieClient(t, c, u)
	return c
}

func maybePersistSession(t *testing.T, reg *driver.RegistryDefault, sess *session.Session) {
	id, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), sess.Identity.ID)
	if err != nil {
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), sess.Identity))
		id, err = reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), sess.Identity.ID)
		require.NoError(t, err)
	}
	sess.Identity = id
	sess.IdentityID = id.ID

	require.NoError(t, err, reg.SessionPersister().UpsertSession(context.Background(), sess))
}

func NewHTTPClientWithSessionCookie(t *testing.T, reg *driver.RegistryDefault, sess *session.Session) *http.Client {
	maybePersistSession(t, reg, sess)

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, reg.SessionManager().IssueCookie(context.Background(), w, r, sess))
	})

	if _, ok := reg.CSRFHandler().(*nosurf.CSRFHandler); ok {
		handler = nosurf.New(handler)
	}

	ts := httptest.NewServer(handler)
	defer ts.Close()

	c := NewClientWithCookies(t)
	MockHydrateCookieClient(t, c, ts.URL)
	return c
}

func NewHTTPClientWithSessionCookieLocalhost(t *testing.T, reg *driver.RegistryDefault, sess *session.Session) *http.Client {
	maybePersistSession(t, reg, sess)

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, reg.SessionManager().IssueCookie(context.Background(), w, r, sess))
	})

	if _, ok := reg.CSRFHandler().(*nosurf.CSRFHandler); ok {
		handler = nosurf.New(handler)
	}

	ts := httptest.NewServer(handler)
	defer ts.Close()

	c := NewClientWithCookies(t)

	ts.URL = strings.Replace(ts.URL, "127.0.0.1", "localhost", 1)
	MockHydrateCookieClient(t, c, ts.URL)
	return c
}

func NewNoRedirectHTTPClientWithSessionCookie(t *testing.T, reg *driver.RegistryDefault, sess *session.Session) *http.Client {
	maybePersistSession(t, reg, sess)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, reg.SessionManager().IssueCookie(context.Background(), w, r, sess))
	}))
	defer ts.Close()

	c := NewNoRedirectClientWithCookies(t)

	MockHydrateCookieClient(t, c, ts.URL)
	return c
}

func NewTransportWithLogger(parent http.RoundTripper, t *testing.T) *TransportWithLogger {
	return &TransportWithLogger{
		RoundTripper: parent,
		t:            t,
	}
}

type TransportWithLogger struct {
	http.RoundTripper
	t *testing.T
}

func (ct *TransportWithLogger) RoundTrip(req *http.Request) (*http.Response, error) {
	ct.t.Logf("Made request to: %s", req.URL.String())
	if ct.RoundTripper == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return ct.RoundTripper.RoundTrip(req)
}

func NewHTTPClientWithSessionToken(t *testing.T, reg *driver.RegistryDefault, sess *session.Session) *http.Client {
	maybePersistSession(t, reg, sess)

	return &http.Client{
		Transport: x.NewTransportWithHeader(http.Header{
			"Authorization": {"Bearer " + sess.Token},
		}),
	}
}

func NewHTTPClientWithArbitrarySessionToken(t *testing.T, reg *driver.RegistryDefault) *http.Client {
	req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
	s, err := session.NewActiveSession(req,
		&identity.Identity{ID: x.NewUUID(), State: identity.StateActive},
		NewSessionLifespanProvider(time.Hour),
		time.Now(),
		identity.CredentialsTypePassword,
		identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(t, err, "Could not initialize session from identity.")

	return NewHTTPClientWithSessionToken(t, reg, s)
}

func NewHTTPClientWithArbitrarySessionCookie(t *testing.T, reg *driver.RegistryDefault) *http.Client {
	req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
	s, err := session.NewActiveSession(req,
		&identity.Identity{ID: x.NewUUID(), State: identity.StateActive},
		NewSessionLifespanProvider(time.Hour),
		time.Now(),
		identity.CredentialsTypePassword,
		identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(t, err, "Could not initialize session from identity.")

	return NewHTTPClientWithSessionCookie(t, reg, s)
}

func NewNoRedirectHTTPClientWithArbitrarySessionCookie(t *testing.T, reg *driver.RegistryDefault) *http.Client {
	req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
	s, err := session.NewActiveSession(req,
		&identity.Identity{ID: x.NewUUID(), State: identity.StateActive},
		NewSessionLifespanProvider(time.Hour),
		time.Now(),
		identity.CredentialsTypePassword,
		identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(t, err, "Could not initialize session from identity.")

	return NewNoRedirectHTTPClientWithSessionCookie(t, reg, s)
}

func NewHTTPClientWithIdentitySessionCookie(t *testing.T, reg *driver.RegistryDefault, id *identity.Identity) *http.Client {
	req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
	s, err := session.NewActiveSession(req,
		id,
		NewSessionLifespanProvider(time.Hour),
		time.Now(),
		identity.CredentialsTypePassword,
		identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(t, err, "Could not initialize session from identity.")

	return NewHTTPClientWithSessionCookie(t, reg, s)
}

func NewHTTPClientWithIdentitySessionToken(t *testing.T, reg *driver.RegistryDefault, id *identity.Identity) *http.Client {
	req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
	s, err := session.NewActiveSession(req,
		id,
		NewSessionLifespanProvider(time.Hour),
		time.Now(),
		identity.CredentialsTypePassword,
		identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(t, err, "Could not initialize session from identity.")

	return NewHTTPClientWithSessionToken(t, reg, s)
}

func EnsureAAL(t *testing.T, c *http.Client, ts *httptest.Server, aal string, methods ...string) {
	res, err := c.Get(ts.URL + session.RouteWhoami)
	require.NoError(t, err)
	sess := x.MustReadAll(res.Body)
	require.NoError(t, res.Body.Close())
	assert.EqualValues(t, aal, gjson.GetBytes(sess, "authenticator_assurance_level").String())
	for _, method := range methods {
		assert.EqualValues(t, method, gjson.GetBytes(sess, "authentication_methods.#(method=="+method+").method").String())
	}
	assert.Len(t, gjson.GetBytes(sess, "authentication_methods").Array(), 1+len(methods))
}

func NewAuthorizedTransport(t *testing.T, reg *driver.RegistryDefault, sess *session.Session) *x.TransportWithHeader {
	maybePersistSession(t, reg, sess)

	return x.NewTransportWithHeader(http.Header{
		"Authorization": {"Bearer " + sess.Token},
	})
}
