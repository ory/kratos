package session

import (
	"context"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"

	"github.com/ory/herodot"
)

// DefaultSessionCookieName returns the default cookie name for the kratos session.
const DefaultSessionCookieName = "ory_kratos_session"

var (
	// ErrNoActiveSessionFound is returned when no active cookie session could be found in the request.
	ErrNoActiveSessionFound = herodot.ErrUnauthorized.WithError("request does not have a valid authentication session").WithReason("No active session was found in this request.")
)

// Manager handles identity sessions.
type Manager interface {
	// Get retrieves a session from the store.
	Get(ctx context.Context, sid string) (*Session, error)

	// Create adds a session to the store.
	Create(ctx context.Context, s *Session) error

	// Delete removes a session from the store
	Delete(ctx context.Context, sid string) error

	CreateToRequest(context.Context, *identity.Identity, http.ResponseWriter, *http.Request) (*Session, error)

	// SaveToRequest creates an HTTP session using cookies.
	SaveToRequest(context.Context, *Session, http.ResponseWriter, *http.Request) error

	// FetchFromRequest creates an HTTP session using cookies.
	FetchFromRequest(context.Context, http.ResponseWriter, *http.Request) (*Session, error)

	// PurgeFromRequest removes an HTTP session.
	PurgeFromRequest(context.Context, http.ResponseWriter, *http.Request) error
}

type ManagementProvider interface {
	SessionManager() Manager
}

type ManagerHTTP struct {
	m          Manager
	c          Configuration
	cookieName string
	r          Registry
}

func NewManagerHTTP(
	c Configuration,
	r Registry,
	m Manager,
) *ManagerHTTP {
	return &ManagerHTTP{
		c:          c,
		r:          r,
		cookieName: DefaultSessionCookieName,
		m:          m,
	}
}

func (s *ManagerHTTP) WithManager(m Manager) {
	s.m = m
}

func (s *ManagerHTTP) CreateToRequest(ctx context.Context, i *identity.Identity, w http.ResponseWriter, r *http.Request) (*Session, error) {
	p := NewSession(i, r, s.c)
	if err := s.m.Create(ctx, p); err != nil {
		return nil, err
	}

	if err := s.SaveToRequest(ctx, p, w, r); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *ManagerHTTP) SaveToRequest(ctx context.Context, session *Session, w http.ResponseWriter, r *http.Request) error {
	cookie, _ := s.r.CookieManager().Get(r, s.cookieName)
	cookie.Values["sid"] = session.SID
	if err := cookie.Save(r, w); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *ManagerHTTP) FetchFromRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (*Session, error) {
	cookie, err := s.r.CookieManager().Get(r, s.cookieName)
	if err != nil {
		if _, ok := err.(securecookie.Error); ok {
			// If securecookie returns an error, the HMAC is probably invalid. In that case, we really want
			// to remove the cookie from the browser as it is invalid anyways.
			if err := s.PurgeFromRequest(ctx, w, r); err != nil {
				return nil, err
			}
		}

		return nil, errors.WithStack(ErrNoActiveSessionFound.WithDebug(err.Error()))
	}

	sid, ok := cookie.Values["sid"].(string)
	if !ok {
		return nil, errors.WithStack(ErrNoActiveSessionFound)
	}

	se, err := s.m.Get(ctx, sid)
	if err != nil && err.Error() == herodot.ErrNotFound.Error() {
		return nil, errors.WithStack(ErrNoActiveSessionFound)
	} else if err != nil {
		return nil, err
	}

	se.Identity = se.Identity.CopyWithoutCredentials()

	return se, nil
}

func (s *ManagerHTTP) PurgeFromRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	cookie, _ := s.r.CookieManager().Get(r, s.cookieName)
	cookie.Options.MaxAge = -1
	if err := cookie.Save(r, w); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
