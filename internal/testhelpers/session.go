// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gobuffalo/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/nosurf"
	"github.com/ory/x/contextx"
)

type SessionLifespanProvider struct {
	e time.Duration
}

func (p *SessionLifespanProvider) SessionLifespan(context.Context) time.Duration {
	return p.e
}

func NewSessionClient(t *testing.T, u string) *http.Client {
	c := NewClientWithCookies(t)
	MockHydrateCookieClient(t, c, u)
	return c
}

func maybePersistSession(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, sess *session.Session) {
	id, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, sess.Identity.ID)
	if err != nil {
		require.NoError(t, sess.Identity.SetAvailableAAL(ctx, reg.IdentityManager()))
		require.NoError(t, reg.IdentityManager().Create(ctx, sess.Identity))
		id, err = reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, sess.Identity.ID)
		require.NoError(t, err)
	}
	sess.Identity = id
	sess.IdentityID = id.ID

	require.NoError(t, err, reg.SessionPersister().UpsertSession(ctx, sess))
}

func NewHTTPClientWithSessionCookie(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, sess *session.Session) *http.Client {
	maybePersistSession(t, ctx, reg, sess)

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, reg.SessionManager().IssueCookie(ctx, w, r, sess))
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

func NewHTTPClientWithSessionCookieLocalhost(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, sess *session.Session) *http.Client {
	maybePersistSession(t, ctx, reg, sess)

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, reg.SessionManager().IssueCookie(ctx, w, r, sess))
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

func NewNoRedirectHTTPClientWithSessionCookie(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, sess *session.Session) *http.Client {
	maybePersistSession(t, ctx, reg, sess)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, reg.SessionManager().IssueCookie(ctx, w, r, sess))
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

func NewHTTPClientWithSessionToken(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, sess *session.Session) *http.Client {
	maybePersistSession(t, ctx, reg, sess)

	return &http.Client{
		Transport: NewTransportWithHeader(t, http.Header{
			"Authorization": {"Bearer " + sess.Token},
		}),
	}
}

func NewHTTPClientWithArbitrarySessionToken(t *testing.T, ctx context.Context, reg *driver.RegistryDefault) *http.Client {
	return NewHTTPClientWithArbitrarySessionTokenAndTraits(t, ctx, reg, nil)
}

func NewHTTPClientWithArbitrarySessionTokenAndTraits(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, traits identity.Traits) *http.Client {
	req := NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil).WithContext(contextx.WithConfigValue(ctx, "session.lifespan", time.Hour))
	s, err := NewActiveSession(req, reg,
		&identity.Identity{ID: x.NewUUID(), State: identity.StateActive, Traits: traits, NID: x.NewUUID(), SchemaID: "default"},
		time.Now(),
		identity.CredentialsTypePassword,
		identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(t, err, "Could not initialize session from identity.")

	return NewHTTPClientWithSessionToken(t, ctx, reg, s)
}

func NewHTTPClientWithArbitrarySessionCookie(t *testing.T, ctx context.Context, reg *driver.RegistryDefault) *http.Client {
	req := NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
	req = req.WithContext(contextx.WithConfigValue(ctx, "session.lifespan", time.Hour))
	id := x.NewUUID()
	s, err := NewActiveSession(req, reg,
		&identity.Identity{ID: id, State: identity.StateActive, Traits: []byte("{}"), Credentials: map[identity.CredentialsType]identity.Credentials{
			identity.CredentialsTypePassword: {Type: "password", Identifiers: []string{id.String()}, Config: []byte(`{"hashed_password":"$2a$04$zvZz1zV"}`)},
		}},
		time.Now(),
		identity.CredentialsTypePassword,
		identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(t, err, "Could not initialize session from identity.")

	return NewHTTPClientWithSessionCookie(t, ctx, reg, s)
}

func NewNoRedirectHTTPClientWithArbitrarySessionCookie(t *testing.T, ctx context.Context, reg *driver.RegistryDefault) *http.Client {
	req := NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
	req = req.WithContext(contextx.WithConfigValue(ctx, "session.lifespan", time.Hour))
	id := x.NewUUID()
	s, err := NewActiveSession(req, reg,
		&identity.Identity{ID: id, State: identity.StateActive,
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {Type: "password", Identifiers: []string{id.String()}, Config: []byte(`{"hashed_password":"$2a$04$zvZz1zV"}`)},
			}},
		time.Now(),
		identity.CredentialsTypePassword,
		identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(t, err, "Could not initialize session from identity.")

	return NewNoRedirectHTTPClientWithSessionCookie(t, ctx, reg, s)
}

func NewHTTPClientWithIdentitySessionCookie(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, id *identity.Identity) *http.Client {
	req := NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
	req = req.WithContext(contextx.WithConfigValue(ctx, "session.lifespan", time.Hour))
	s, err := NewActiveSession(req, reg,
		id,
		time.Now(),
		identity.CredentialsTypePassword,
		identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(t, err, "Could not initialize session from identity.")

	return NewHTTPClientWithSessionCookie(t, ctx, reg, s)
}

func NewHTTPClientWithIdentitySessionCookieLocalhost(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, id *identity.Identity) *http.Client {
	req := NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
	s, err := NewActiveSession(req, reg,
		id,
		time.Now(),
		identity.CredentialsTypePassword,
		identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(t, err, "Could not initialize session from identity.")

	return NewHTTPClientWithSessionCookieLocalhost(t, ctx, reg, s)
}

func NewHTTPClientWithIdentitySessionToken(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, id *identity.Identity) *http.Client {
	req := NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
	req = req.WithContext(contextx.WithConfigValue(ctx, "session.lifespan", time.Hour))
	s, err := NewActiveSession(req, reg,
		id,
		time.Now(),
		identity.CredentialsTypePassword,
		identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(t, err, "Could not initialize session from identity.")

	return NewHTTPClientWithSessionToken(t, ctx, reg, s)
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

func NewAuthorizedTransport(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, sess *session.Session) *TransportWithHeader {
	maybePersistSession(t, ctx, reg, sess)

	return NewTransportWithHeader(t, http.Header{
		"Authorization": {"Bearer " + sess.Token},
	})
}

func NewTransportWithHeader(t *testing.T, h http.Header) *TransportWithHeader {
	if t == nil {
		panic("This function is for testing use only.")
	}
	return &TransportWithHeader{
		RoundTripper: http.DefaultTransport,
		h:            h,
	}
}

type TransportWithHeader struct {
	http.RoundTripper
	h http.Header
}

func (ct *TransportWithHeader) GetHeader() http.Header {
	return ct.h
}

func (ct *TransportWithHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	for k := range ct.h {
		req.Header.Set(k, ct.h.Get(k))
	}
	return ct.RoundTripper.RoundTrip(req)
}

func AssertNoCSRFCookieInResponse(t *testing.T, _ *httptest.Server, _ *http.Client, r *http.Response) {
	found := false
	for _, c := range r.Cookies() {
		if strings.HasPrefix(c.Name, "csrf_token") {
			found = true
		}
	}
	require.False(t, found)
}
