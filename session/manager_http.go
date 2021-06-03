package session

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"

	"github.com/ory/x/sqlcon"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

type (
	managerHTTPDependencies interface {
		config.Provider
		identity.PoolProvider
		x.CookieProvider
		x.CSRFProvider
		PersistenceProvider
	}
	ManagerHTTP struct {
		cookieName func(ctx context.Context) string
		r          managerHTTPDependencies
	}
)

func NewManagerHTTP(r managerHTTPDependencies) *ManagerHTTP {
	return &ManagerHTTP{
		r: r,
		cookieName: func(ctx context.Context) string {
			return r.Config(ctx).SessionName()
		},
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
	cookie, _ := s.r.CookieManager(r.Context()).Get(r, s.cookieName(ctx))

	if domain := s.r.Config(ctx).SessionDomain(); domain != "" {
		cookie.Options.Domain = domain
	} else if alias := s.r.Config(ctx).SelfPublicURL(r); s.r.Config(ctx).SelfPublicURL(nil).String() != alias.String() {
		// If a domain alias is detected use that instead.
		cookie.Options.Domain = alias.Hostname()
		cookie.Options.Path = alias.Path
	}

	old, err := s.FetchFromRequest(ctx, r)
	if err != nil {
		// No session was set prior -> regenerate anti-csrf token
		_ = s.r.CSRFHandler().RegenerateToken(w, r)
	} else if old.Identity.ID != session.Identity.ID {
		// No session was set prior -> regenerate anti-csrf token
		_ = s.r.CSRFHandler().RegenerateToken(w, r)
	}

	if s.r.Config(ctx).SessionPath() != "" {
		cookie.Options.Path = s.r.Config(ctx).SessionPath()
	}

	if s.r.Config(ctx).SessionSameSiteMode() != 0 {
		cookie.Options.SameSite = s.r.Config(ctx).SessionSameSiteMode()
	}

	cookie.Options.MaxAge = 0
	if s.r.Config(ctx).SessionPersistentCookie() {
		cookie.Options.MaxAge = int(s.r.Config(ctx).SessionLifespan().Seconds())
	}

	cookie.Values["session_token"] = session.Token
	if err := cookie.Save(r, w); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *ManagerHTTP) extractToken(r *http.Request) string {
	if token := r.Header.Get("X-Session-Token"); len(token) > 0 {
		return token
	}

	cookie, err := s.r.CookieManager(r.Context()).Get(r, s.cookieName(r.Context()))
	if err != nil {
		token, _ := bearerTokenFromRequest(r)
		return token
	}

	token, ok := cookie.Values["session_token"].(string)
	if ok {
		return token
	}

	token, _ = bearerTokenFromRequest(r)
	return token
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

	cookie, _ := s.r.CookieManager(r.Context()).Get(r, s.cookieName(ctx))
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
