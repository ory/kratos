package testhelpers

import (
	"context"
	"net/http"
	"testing"
	"time"

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

func (p *SessionLifespanProvider) SessionLifespan() time.Duration {
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

	require.NoError(t, err, reg.SessionPersister().CreateSession(context.Background(), sess))
}

func NewHTTPClientWithSessionCookie(t *testing.T, reg *driver.RegistryDefault, sess *session.Session) *http.Client {
	maybePersistSession(t, reg, sess)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, reg.SessionManager().IssueCookie(context.Background(), w, r, sess))
	}))
	defer ts.Close()

	c := NewClientWithCookies(t)

	// This should work for other test servers as well because cookies ignore ports.
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
	return NewHTTPClientWithSessionToken(t, reg, session.NewActiveSession(
		&identity.Identity{ID: x.NewUUID()},
		NewSessionLifespanProvider(time.Hour),
		time.Now(),
	))
}

func NewHTTPClientWithArbitrarySessionCookie(t *testing.T, reg *driver.RegistryDefault) *http.Client {
	return NewHTTPClientWithSessionCookie(t, reg, session.NewActiveSession(
		&identity.Identity{ID: x.NewUUID()},
		NewSessionLifespanProvider(time.Hour),
		time.Now(),
	))
}

func NewHTTPClientWithIdentitySessionCookie(t *testing.T, reg *driver.RegistryDefault, id *identity.Identity) *http.Client {
	return NewHTTPClientWithSessionCookie(t, reg,
		session.NewActiveSession(id, NewSessionLifespanProvider(time.Hour), time.Now()))
}

func NewHTTPClientWithIdentitySessionToken(t *testing.T, reg *driver.RegistryDefault, id *identity.Identity) *http.Client {
	return NewHTTPClientWithSessionToken(t, reg,
		session.NewActiveSession(id, NewSessionLifespanProvider(time.Hour), time.Now()))
}
