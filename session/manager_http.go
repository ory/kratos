package session

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/pkg/errors"

	"github.com/ory/x/sqlcon"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

type (
	managerHTTPDependencies interface {
		PersistenceProvider
		x.CookieProvider
		identity.PoolProvider
		x.CSRFProvider
	}
	managerHTTPConfiguration interface {
		SessionLifespan() time.Duration
		SessionSecrets() [][]byte
	}
	ManagerHTTP struct {
		c          managerHTTPConfiguration
		cookieName string
		r          managerHTTPDependencies
	}
)

func NewManagerHTTP(
	c managerHTTPConfiguration,
	r managerHTTPDependencies,
) *ManagerHTTP {
	return &ManagerHTTP{
		c:          c,
		r:          r,
		cookieName: DefaultSessionCookieName,
	}
}

func (s *ManagerHTTP) CreateToRequest(ctx context.Context, i *identity.Identity, w http.ResponseWriter, r *http.Request) (*Session, error) {
	p := NewSession(i, r, s.c)
	if err := s.r.SessionPersister().CreateSession(ctx, p); err != nil {
		return nil, err
	}

	if err := s.SaveToRequest(ctx, p, w, r); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *ManagerHTTP) SaveToRequest(ctx context.Context, session *Session, w http.ResponseWriter, r *http.Request) error {
	_ = s.r.CSRFHandler().RegenerateToken(w, r)
	cookie, _ := s.r.CookieManager().Get(r, s.cookieName)
	cookie.Values["sid"] = session.ID.String()
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

	se, err := s.r.SessionPersister().GetSession(ctx, x.ParseUUID(sid))
	if err != nil && (err.Error() == herodot.ErrNotFound.Error() ||
		err.Error() == sqlcon.ErrNoRows.Error()) {
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
