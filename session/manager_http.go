package session

import (
	"context"
	"net/http"
	"time"

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
		SecretsSession() [][]byte
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

func (s *ManagerHTTP) CreateToRequest(ctx context.Context, w http.ResponseWriter, r *http.Request, ss *Session) error {
	if err := s.r.SessionPersister().CreateSession(ctx, ss); err != nil {
		return err
	}

	if err := s.SaveToRequest(ctx, w, r, ss); err != nil {
		return err
	}

	return nil
}

func (s *ManagerHTTP) SaveToRequest(ctx context.Context, w http.ResponseWriter, r *http.Request, session *Session) error {
	_ = s.r.CSRFHandler().RegenerateToken(w, r)
	cookie, _ := s.r.CookieManager().Get(r, s.cookieName)
	cookie.Values["sid"] = session.ID.String()
	if err := cookie.Save(r, w); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *ManagerHTTP) FetchFromRequest(ctx context.Context, r *http.Request) (*Session, error) {
	cookie, err := s.r.CookieManager().Get(r, s.cookieName)
	if err != nil {
		return nil, errors.WithStack(ErrNoActiveSessionFound.WithWrap(err).WithDebugf("%s", err))
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
