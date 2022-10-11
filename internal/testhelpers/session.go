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

type SessionPrivilegedMaxAgeProvider struct {
	e time.Duration
}

func (p *SessionPrivilegedMaxAgeProvider) SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx context.Context) time.Duration {
	return p.e
}

func NewSessionPrivilegedMaxAgeProvider(privilegedMaxAge time.Duration) *SessionPrivilegedMaxAgeProvider {
	return &SessionPrivilegedMaxAgeProvider{e: privilegedMaxAge}
}

type SessionLifespanAndPrivilegedMaxAgeProvider struct {
	l *SessionLifespanProvider
	p *SessionPrivilegedMaxAgeProvider
}

func (p *SessionLifespanAndPrivilegedMaxAgeProvider) SessionLifespan(ctx context.Context) time.Duration {
	return p.l.SessionLifespan(ctx)
}

func (p *SessionLifespanAndPrivilegedMaxAgeProvider) SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx context.Context) time.Duration {
	return p.p.SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx)
}

func NewSessionLifespanAndPrivilegedMaxAgeProvider(expiresIn time.Duration, privilegedMaxAge time.Duration) *SessionLifespanAndPrivilegedMaxAgeProvider {
	return &SessionLifespanAndPrivilegedMaxAgeProvider{
		l: NewSessionLifespanProvider(expiresIn),
		p: NewSessionPrivilegedMaxAgeProvider(privilegedMaxAge),
	}
}

// Gets a session provider that is privileged
func PrivilegedProvider() *SessionLifespanAndPrivilegedMaxAgeProvider {
	return NewSessionLifespanAndPrivilegedMaxAgeProvider(time.Hour, time.Minute)
}

// Gets a session provider that is unprivileged
func UnprivilegedProvider() *SessionLifespanAndPrivilegedMaxAgeProvider {
	return NewSessionLifespanAndPrivilegedMaxAgeProvider(time.Hour, time.Nanosecond)
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

// func NewHTTPClientWithArbitrarySessionToken(t *testing.T, ctx context.Context, reg *driver.RegistryDefault) *http.Client {
// 	req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
// 	privilegedMaxAge := reg.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx)
// 	s, err := session.NewActiveSession(req,
// 		&identity.Identity{ID: x.NewUUID(), State: identity.StateActive},
// 		NewSessionLifespanAndPrivilegedMaxAgeProvider(time.Hour, privilegedMaxAge),
// 		time.Now(),
// 		identity.CredentialsTypePassword,
// 		identity.AuthenticatorAssuranceLevel1,
// 	)
// 	require.NoError(t, err, "Could not initialize session from identity.")

// 	return NewHTTPClientWithSessionToken(t, reg, s)
// }

// func NewHTTPClientWithArbitrarySessionCookie(t *testing.T, ctx context.Context, reg *driver.RegistryDefault) *http.Client {
// 	req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
// 	privilegedMaxAge := reg.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx)
// 	s, err := session.NewActiveSession(req,
// 		&identity.Identity{ID: x.NewUUID(), State: identity.StateActive},
// 		NewSessionLifespanAndPrivilegedMaxAgeProvider(time.Hour, privilegedMaxAge),
// 		time.Now(),
// 		identity.CredentialsTypePassword,
// 		identity.AuthenticatorAssuranceLevel1,
// 	)
// 	require.NoError(t, err, "Could not initialize session from identity.")

// 	return NewHTTPClientWithSessionCookie(t, reg, s)
// }

// func NewNoRedirectHTTPClientWithArbitrarySessionCookie(t *testing.T, ctx context.Context, reg *driver.RegistryDefault) *http.Client {
// 	req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
// 	privilegedMaxAge := reg.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx)
// 	s, err := session.NewActiveSession(req,
// 		&identity.Identity{ID: x.NewUUID(), State: identity.StateActive},
// 		NewSessionLifespanAndPrivilegedMaxAgeProvider(time.Hour, privilegedMaxAge),
// 		time.Now(),
// 		identity.CredentialsTypePassword,
// 		identity.AuthenticatorAssuranceLevel1,
// 	)
// 	require.NoError(t, err, "Could not initialize session from identity.")

// 	return NewNoRedirectHTTPClientWithSessionCookie(t, reg, s)
// }

// func NewHTTPClientWithIdentitySessionCookie(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, id *identity.Identity) *http.Client {
// 	req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
// 	privilegedMaxAge := reg.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx)
// 	s, err := session.NewActiveSession(req,
// 		id,
// 		NewSessionLifespanAndPrivilegedMaxAgeProvider(time.Hour, privilegedMaxAge),
// 		time.Now(),
// 		identity.CredentialsTypePassword,
// 		identity.AuthenticatorAssuranceLevel1,
// 	)
// 	require.NoError(t, err, "Could not initialize session from identity.")

// 	return NewHTTPClientWithSessionCookie(t, reg, s)
// }

// func NewHTTPClientWithIdentitySessionToken(t *testing.T, ctx context.Context, reg *driver.RegistryDefault, id *identity.Identity) *http.Client {
// 	req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
// 	privilegedMaxAge := reg.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx)
// 	s, err := session.NewActiveSession(req,
// 		id,
// 		NewSessionLifespanAndPrivilegedMaxAgeProvider(time.Hour, privilegedMaxAge),
// 		time.Now(),
// 		identity.CredentialsTypePassword,
// 		identity.AuthenticatorAssuranceLevel1,
// 	)
// 	require.NoError(t, err, "Could not initialize session from identity.")

// 	return NewHTTPClientWithSessionToken(t, reg, s)
// }

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

// A builder struct to help customize an authenticated http client
type HTTPClientBuilder struct {
	t        *testing.T
	request  *http.Request
	identity *identity.Identity
	session  *session.Session
}

func NewHTTPClientBuilder(t *testing.T) *HTTPClientBuilder {
	return &HTTPClientBuilder{t: t}
}

// Raise a test error if the request hasn't been assigned yet.
func (c *HTTPClientBuilder) nilRequestCheck() {
	if c.request == nil {
		c.t.Errorf("creating a Session requires a request to be present on the builder, but request was nil")
	}
}

// Raise a test error if the identity hasn't been assigned yet.
func (c *HTTPClientBuilder) nilIdentityCheck() {
	if c.request == nil {
		c.t.Errorf("creating a Session requires an identity to be present on the builder, but identity was nil")
	}
}

// Raise a test error if the session hasn't been assigned yet.
func (c *HTTPClientBuilder) nilSessionCheck() {
	if c.session == nil {
		c.t.Errorf("consuming the HTTPClientBuilder requires a session to be present on the builder, but session was nil")
	}
}

// Sets a custom request to be used during creation of the session.
//
// A request is required unless you intend to provide your own Session via `SetSession`.
func (c *HTTPClientBuilder) SetReqest(req *http.Request) *HTTPClientBuilder {
	c.request = req
	return c
}

func (c *HTTPClientBuilder) SetReqestFromWhoAmI() *HTTPClientBuilder {
	return c.SetReqest(
		x.NewTestHTTPRequest(c.t, "GET", "/sessions/whoami", nil),
	)
}

// Sets a given identity on the builder struct for later use in instantiating the session.
//
// An identity is required unless you intend to provide your own Session via `SetSession`.
func (c *HTTPClientBuilder) SetIdentity(identity *identity.Identity) *HTTPClientBuilder {
	c.identity = identity
	return c
}

// Using this builder method for setting the identity is equivalent to using the old "WithArbitrary"
// functions, since this creates a new arbitrary session token for the session.
//
// An identity is required unless you intend to provide your own Session via `SetSession`.
func (c *HTTPClientBuilder) SetIdentityFromNew() *HTTPClientBuilder {
	return c.SetIdentity(
		&identity.Identity{ID: x.NewUUID(), State: identity.StateActive},
	)
}

// Sets a given session on the builder for instantiating the final client
func (c *HTTPClientBuilder) SetSession(session *session.Session) *HTTPClientBuilder {
	c.session = session
	return c
}

// Sets a default session on the builder. This session is active, has an authenticated_at time of now,
// expires in 1 hour, is privileged for 1 minute, has a password credential type, and is aal1
func (c *HTTPClientBuilder) SetSessionDefault() *HTTPClientBuilder {
	c.nilRequestCheck()
	c.nilIdentityCheck()
	session, err := session.NewActiveSession(
		c.request,
		c.identity,
		NewSessionLifespanAndPrivilegedMaxAgeProvider(time.Hour, time.Minute),
		time.Now(),
		identity.CredentialsTypePassword,
		identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(c.t, err, "Could not initialize session from identity.")
	return c.SetSession(session)
}

// Sets a new session based on a default session, but using a provided instance of
// `SessionLifespanAndPrivilegedMaxAgeProvider` for setting the session lifespan and privileged age.
//
// Example Usage:
//
//	func TestSessionBehavior(t *testing.T) {
//		var unprivilegedClient *http.Client
//		unprivilegedClient = NewHTTPClientBuilder(t).
//			SetReqestFromWhoAmI().
//			SetIdentityFromNew().
//			SetSessionWithProvider(UnrivilegedProvider()).
//			ClientWithSessionToken()
//	}
func (c *HTTPClientBuilder) SetSessionWithProvider(provider *SessionLifespanAndPrivilegedMaxAgeProvider) *HTTPClientBuilder {
	c.nilRequestCheck()
	c.nilIdentityCheck()
	session, err := session.NewActiveSession(
		c.request,
		c.identity,
		provider,
		time.Now(),
		identity.CredentialsTypePassword,
		identity.AuthenticatorAssuranceLevel1,
	)
	require.NoError(c.t, err, "Could not initialize session from identity.")
	return c.SetSession(session)
}

// Record a callback function for mutating the session after the client has been created. This
// is more useful than recording the session pointer because mutations to the session wouldn't
// be persistend unless the session was then passed to the private `testhelpers.maybePersistSession`
// function. This function compromises in giving access to the session without uncontrollably
// exposing the persistence function.
//
// Example Usage:
//
//	func TestSessionBehavior(t *testing.T) {
//		// Define a var at which we'll record the session mutation function
//		var sessionMutator func(mutateFunc func(session *session.Session))
//
//		// Construct a client with a default unprivileged session
//		client := NewHTTPClientBuilder(t).
//			SetReqestFromWhoAmI().
//			SetIdentityFromNew().
//			SetSessionDefault(UnrivilegedProvider()).
//			RecordSessionMutator(reg, &sessionMutator).
//			ClientWithSessionToken()
//
//		// Run some tests using `client` requiring the session to be unprivileged
//		...
//
//		// Pretend the session has been refreshed, and is now privileged for an additional minute
//		sessionMutator(func(session *session.Session) {
//			session.SetPrivilegedUntil(time.Now().Add(time.Minute))
//		})
//
//		// Run some tests using `client` requiring the session to be privileged
//		...
//	}
func (c *HTTPClientBuilder) RecordSessionMutator(reg *driver.RegistryDefault, funcRef *func(func(session *session.Session))) *HTTPClientBuilder {
	*funcRef = func(mutateFunc func(session *session.Session)) {
		mutateFunc(c.session)
		maybePersistSession(c.t, reg, c.session)
	}
	return c
}

// Builder Consumers

// A generic function for creating a new client using the builders session.
func (c *HTTPClientBuilder) ClientCallback(clientFunc func(*testing.T, *session.Session) *http.Client) *http.Client {
	return clientFunc(c.t, c.session)
}

// A quality of life function that consumes the builder, creating a client with a session token
// via the `NewHTTPClientWithSessionToken` function.
func (c *HTTPClientBuilder) ClientWithSessionToken(reg *driver.RegistryDefault) *http.Client {
	c.nilSessionCheck()
	return c.ClientCallback(func(t *testing.T, s *session.Session) *http.Client {
		return NewHTTPClientWithSessionToken(t, reg, s)
	})
}

// A quality of life function that consumes the builder, creating a client with a session cookie
// via the `NewHTTPClientWithSessionCookie` function.
func (c *HTTPClientBuilder) ClientWithSessionCookie(reg *driver.RegistryDefault) *http.Client {
	c.nilSessionCheck()
	return c.ClientCallback(func(t *testing.T, s *session.Session) *http.Client {
		return NewHTTPClientWithSessionCookie(t, reg, s)
	})
}

// A quality of life function that consumes the builder, creating a client with a session cookie
// without redirections via the `NewNoRedirectHTTPClientWithSessionCookie` function.
func (c *HTTPClientBuilder) ClientNoRedirectWithSessionCookie(reg *driver.RegistryDefault) *http.Client {
	c.nilSessionCheck()
	return c.ClientCallback(func(t *testing.T, s *session.Session) *http.Client {
		return NewNoRedirectHTTPClientWithSessionCookie(t, reg, s)
	})
}

// Common configurations

func NewIdentityClientWithSessionToken(t *testing.T, reg *driver.RegistryDefault, id *identity.Identity) *http.Client {
	return NewHTTPClientBuilder(t).
		SetReqestFromWhoAmI().
		SetIdentity(id).
		SetSessionDefault().
		ClientWithSessionToken(reg)
}

func NewIdentityClientWithSessionCookie(t *testing.T, reg *driver.RegistryDefault, id *identity.Identity) *http.Client {
	return NewHTTPClientBuilder(t).
		SetReqestFromWhoAmI().
		SetIdentity(id).
		SetSessionDefault().
		ClientWithSessionToken(reg)
}

func NewDefaultClientWithSessionToken(t *testing.T, reg *driver.RegistryDefault) *http.Client {
	return NewHTTPClientBuilder(t).
		SetReqestFromWhoAmI().
		SetIdentityFromNew().
		SetSessionDefault().
		ClientWithSessionToken(reg)
}

func NewDefaultClientWithSessionCookie(t *testing.T, reg *driver.RegistryDefault) *http.Client {
	return NewHTTPClientBuilder(t).
		SetReqestFromWhoAmI().
		SetIdentityFromNew().
		SetSessionDefault().
		ClientWithSessionCookie(reg)
}

func NewDefaultClientNoRedirectWithSessionCookie(t *testing.T, reg *driver.RegistryDefault) *http.Client {
	return NewHTTPClientBuilder(t).
		SetReqestFromWhoAmI().
		SetIdentityFromNew().
		SetSessionDefault().
		ClientNoRedirectWithSessionCookie(reg)
}

func NewDefaultClientViaCallback(t *testing.T, clientFunc func(*testing.T, *session.Session) *http.Client) *http.Client {
	return NewHTTPClientBuilder(t).
		SetReqestFromWhoAmI().
		SetIdentityFromNew().
		SetSessionDefault().
		ClientCallback(clientFunc)
}
