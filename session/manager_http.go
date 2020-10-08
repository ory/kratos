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
		SessionPersistentCookie() bool
		SessionLifespan() time.Duration
		SecretsSession() [][]byte
		SessionSameSiteMode() http.SameSite
		SessionDomain() string
		SessionPath() string
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

func (s *ManagerHTTP) CreateAndIssueCookie(ctx context.Context, w http.ResponseWriter, r *http.Request, ss *Session) error {
	if err := s.r.SessionPersister().CreateSession(ctx, ss); err != nil {
		return err
	}

	if err := s.IssueCookie(ctx, w, r, ss); err != nil {
		return err
	}

	return nil
}

func (s *ManagerHTTP) IssueCookie(ctx context.Context, w http.ResponseWriter, r *http.Request, session *Session) error {
	cookie, _ := s.r.CookieManager().Get(r, s.cookieName)
	if s.c.SessionDomain() != "" {
		cookie.Options.Domain = s.c.SessionDomain()
	}

	old, err := s.FetchFromRequest(context.Background(), r)
	if err != nil {
		// No session was set prior -> regenerate anti-csrf token
		_ = s.r.CSRFHandler().RegenerateToken(w, r)
	} else if old.Identity.ID != session.Identity.ID {
		// No session was set prior -> regenerate anti-csrf token
		_ = s.r.CSRFHandler().RegenerateToken(w, r)
	}

	if s.c.SessionPath() != "" {
		cookie.Options.Path = s.c.SessionPath()
	}

	if s.c.SessionSameSiteMode() != 0 {
		cookie.Options.SameSite = s.c.SessionSameSiteMode()
	}

	cookie.Options.MaxAge = 0
	if s.c.SessionPersistentCookie() {
		cookie.Options.MaxAge = int(s.c.SessionLifespan().Seconds())
	}

	cookie.Values["session_token"] = session.Token
	if err := cookie.Save(r, w); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *ManagerHTTP) extractToken(r *http.Request) string {
	if token, ok := bearerTokenFromRequest(r); ok {
		return token
	}

	if token := r.Header.Get("X-Session-Token"); len(token) > 0 {
		return token
	}

	cookie, err := s.r.CookieManager().Get(r, s.cookieName)
	if err != nil {
		return ""
	}

	token, ok := cookie.Values["session_token"].(string)
	if ok {
		return token
	}

	return ""
}

func (s *ManagerHTTP) FetchFromRequest(ctx context.Context, r *http.Request) (*Session, error) {
	token := s.extractToken(r)
	if token == "" {
		return nil, errors.WithStack(ErrNoActiveSessionFound)
	}

	se, err := s.r.SessionPersister().GetSessionByToken(ctx, token)
	if err != nil {
		if errors.Is(err, herodot.ErrNotFound) || errors.Is(err, sqlcon.ErrNoRows) {
			return nil, errors.WithStack(ErrNoActiveSessionFound)
		}
		return nil, err
	}

	if !se.IsActive() {
		return nil, errors.WithStack(ErrNoActiveSessionFound)
	}

	se.Identity = se.Identity.CopyWithoutCredentials()
	return se, nil
}

func (s *ManagerHTTP) PurgeFromRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if token, ok := bearerTokenFromRequest(r); ok {
		return errors.WithStack(s.r.SessionPersister().RevokeSessionByToken(ctx, token))
	}

	cookie, _ := s.r.CookieManager().Get(r, s.cookieName)
	token, ok := cookie.Values["session_token"].(string)
	if !ok {
		return nil
	}

	if err := s.r.SessionPersister().RevokeSessionByToken(ctx, token); err != nil {
		return errors.WithStack(err)
	}

	cookie.Options.MaxAge = -1
	if err := cookie.Save(r, w); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
