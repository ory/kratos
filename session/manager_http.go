package session

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/x/randx"

	"github.com/gorilla/sessions"

	"github.com/ory/x/urlx"

	"github.com/gofrs/uuid"

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
		identity.PrivilegedPoolProvider
		identity.ManagementProvider
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
			return r.Config().SessionName(ctx)
		},
	}
}

func (s *ManagerHTTP) UpsertAndIssueCookie(ctx context.Context, w http.ResponseWriter, r *http.Request, ss *Session) error {
	if err := s.r.SessionPersister().UpsertSession(ctx, ss); err != nil {
		return err
	}

	if err := s.IssueCookie(ctx, w, r, ss); err != nil {
		return err
	}

	return nil
}

func (s *ManagerHTTP) RefreshCookie(ctx context.Context, w http.ResponseWriter, r *http.Request, session *Session) error {
	// If it is a session token there is nothing to do.
	cookieHeader := r.Header.Get("X-Session-Cookie")
	_, cookieErr := r.Cookie(s.cookieName(r.Context()))
	if len(cookieHeader) == 0 && errors.Is(cookieErr, http.ErrNoCookie) {
		return nil
	}

	cookie, err := s.getCookie(r)
	if err != nil {
		return err
	}

	expiresAt := getCookieExpiry(cookie)
	if expiresAt == nil || expiresAt.Before(session.ExpiresAt) {
		if err := s.IssueCookie(ctx, w, r, session); err != nil {
			return err
		}
	}

	return nil
}

func (s *ManagerHTTP) IssueCookie(ctx context.Context, w http.ResponseWriter, r *http.Request, session *Session) error {
	cookie, err := s.r.CookieManager(r.Context()).Get(r, s.cookieName(ctx))
	// Fix for https://github.com/ory/kratos/issues/1695
	if err != nil && cookie == nil {
		return errors.WithStack(err)
	}

	if s.r.Config().SessionPath(ctx) != "" {
		cookie.Options.Path = s.r.Config().SessionPath(ctx)
	}

	if domain := s.r.Config().SessionDomain(ctx); domain != "" {
		cookie.Options.Domain = domain
	}

	if alias := s.r.Config().SelfPublicURL(ctx); s.r.Config().SelfPublicURL(ctx).String() != alias.String() {
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

	if s.r.Config().SessionSameSiteMode(ctx) != 0 {
		cookie.Options.SameSite = s.r.Config().SessionSameSiteMode(ctx)
	}

	cookie.Options.MaxAge = 0
	if s.r.Config().SessionPersistentCookie(ctx) {
		if session.ExpiresAt.IsZero() {
			cookie.Options.MaxAge = int(s.r.Config().SessionLifespan(ctx).Seconds())
		} else {
			cookie.Options.MaxAge = int(time.Until(session.ExpiresAt).Seconds())
		}
	}

	cookie.Values["session_token"] = session.Token
	cookie.Values["expires_at"] = session.ExpiresAt.UTC().Format(time.RFC3339Nano)
	cookie.Values["nonce"] = randx.MustString(8, randx.Alpha) // Guarantee new kratos session identifier

	if err := cookie.Save(r, w); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func getCookieExpiry(s *sessions.Session) *time.Time {
	expiresAt, ok := s.Values["expires_at"].(string)
	if !ok {
		return nil
	}

	n, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return nil
	}
	return &n
}

func (s *ManagerHTTP) getCookie(r *http.Request) (*sessions.Session, error) {
	if cookie := r.Header.Get("X-Session-Cookie"); len(cookie) > 0 {
		rr := *r
		r = &rr
		r.Header = http.Header{"Cookie": []string{s.cookieName(r.Context()) + "=" + cookie}}
	}

	return s.r.CookieManager(r.Context()).Get(r, s.cookieName(r.Context()))
}

func (s *ManagerHTTP) extractToken(r *http.Request) string {
	if token := r.Header.Get("X-Session-Token"); len(token) > 0 {
		return token
	}

	cookie, err := s.getCookie(r)
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
		return nil, errors.WithStack(NewErrNoCredentialsForSession())
	}

	se, err := s.r.SessionPersister().GetSessionByToken(ctx, token, ExpandEverything)
	if err != nil {
		if errors.Is(err, herodot.ErrNotFound) || errors.Is(err, sqlcon.ErrNoRows) {
			return nil, errors.WithStack(NewErrNoActiveSessionFound())
		}
		return nil, err
	}

	if !se.IsActive() {
		return nil, errors.WithStack(NewErrNoActiveSessionFound())
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

func (s *ManagerHTTP) DoesSessionSatisfy(r *http.Request, sess *Session, requestedAAL string) error {
	sess.SetAuthenticatorAssuranceLevel()
	switch requestedAAL {
	case string(identity.AuthenticatorAssuranceLevel1):
		if sess.AuthenticatorAssuranceLevel >= identity.AuthenticatorAssuranceLevel1 {
			return nil
		}
	case config.HighestAvailableAAL:
		i, err := s.r.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), sess.IdentityID)
		if err != nil {
			return err
		}

		available := identity.NoAuthenticatorAssuranceLevel
		if firstCount, err := s.r.IdentityManager().CountActiveFirstFactorCredentials(r.Context(), i); err != nil {
			return err
		} else if firstCount > 0 {
			available = identity.AuthenticatorAssuranceLevel1
		}

		if secondCount, err := s.r.IdentityManager().CountActiveMultiFactorCredentials(r.Context(), i); err != nil {
			return err
		} else if secondCount > 0 {
			available = identity.AuthenticatorAssuranceLevel2
		}

		if sess.AuthenticatorAssuranceLevel >= available {
			return nil
		}

		return NewErrAALNotSatisfied(
			urlx.CopyWithQuery(urlx.AppendPaths(s.r.Config().SelfPublicURL(r.Context()), "/self-service/login/browser"), url.Values{"aal": {"aal2"}}).String())
	}
	return errors.Errorf("requested unknown aal: %s", requestedAAL)
}

func (s *ManagerHTTP) SessionAddAuthenticationMethods(ctx context.Context, sid uuid.UUID, ams ...AuthenticationMethod) error {
	// Since we added the method, it also means that we have authenticated it
	sess, err := s.r.SessionPersister().GetSession(ctx, sid, ExpandNothing)
	if err != nil {
		return err
	}
	for _, m := range ams {
		sess.CompletedLoginFor(m.Method, m.AAL)
	}
	sess.SetAuthenticatorAssuranceLevel()
	return s.r.SessionPersister().UpsertSession(ctx, sess)
}
