// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/x/events"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/otelx"

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
		x.TracingProvider
		PersistenceProvider
		sessiontokenexchange.PersistenceProvider
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

type options struct {
	requestURL string
}

type ManagerOptions func(*options)

// WithRequestURL passes along query parameters from the requestURL to the new URL (if any exist)
func WithRequestURL(requestURL string) ManagerOptions {
	return func(opts *options) {
		opts.requestURL = requestURL
	}
}

func (s *ManagerHTTP) UpsertAndIssueCookie(ctx context.Context, w http.ResponseWriter, r *http.Request, ss *Session) (err error) {
	ctx, span := s.r.Tracer(ctx).Tracer().Start(ctx, "sessions.ManagerHTTP.UpsertAndIssueCookie")
	defer otelx.End(span, &err)

	isNew := ss.ID == uuid.Nil
	if err := s.r.SessionPersister().UpsertSession(ctx, ss); err != nil {
		return err
	}

	if err := s.IssueCookie(ctx, w, r, ss); err != nil {
		return err
	}

	var event = events.NewSessionChanged
	if isNew {
		event = events.NewSessionIssued
	}

	trace.SpanFromContext(r.Context()).AddEvent(event(r.Context(), string(ss.AuthenticatorAssuranceLevel), ss.ID, ss.IdentityID))
	return nil
}

func (s *ManagerHTTP) RefreshCookie(ctx context.Context, w http.ResponseWriter, r *http.Request, session *Session) (err error) {
	ctx, span := s.r.Tracer(ctx).Tracer().Start(ctx, "sessions.ManagerHTTP.RefreshCookie")
	defer otelx.End(span, &err)

	// If it is a session token there is nothing to do.
	_, cookieErr := r.Cookie(s.cookieName(r.Context()))
	if errors.Is(cookieErr, http.ErrNoCookie) {
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

func (s *ManagerHTTP) IssueCookie(ctx context.Context, w http.ResponseWriter, r *http.Request, session *Session) (err error) {
	ctx, span := s.r.Tracer(ctx).Tracer().Start(ctx, "sessions.ManagerHTTP.IssueCookie")
	defer otelx.End(span, &err)

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
	return s.r.CookieManager(r.Context()).Get(r, s.cookieName(r.Context()))
}

func (s *ManagerHTTP) extractToken(r *http.Request) string {
	_, span := s.r.Tracer(r.Context()).Tracer().Start(r.Context(), "sessions.ManagerHTTP.extractToken")
	defer span.End()

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

func (s *ManagerHTTP) FetchFromRequest(ctx context.Context, r *http.Request) (_ *Session, err error) {
	ctx, span := s.r.Tracer(ctx).Tracer().Start(ctx, "sessions.ManagerHTTP.FetchFromRequest")
	defer func() {
		if e := new(ErrNoActiveSessionFound); errors.As(err, &e) {
			span.End()
		} else {
			otelx.End(span, &err)
		}
	}()

	token := s.extractToken(r.WithContext(ctx))
	if token == "" {
		return nil, errors.WithStack(NewErrNoCredentialsForSession())
	}

	expand := identity.ExpandDefault
	if s.r.Config().SessionWhoAmIAAL(r.Context()) == config.HighestAvailableAAL {
		// When the session endpoint requires the highest AAL, we fetch all credentials immediately to save a
		// query later in "DoesSessionSatisfy". This is a SQL optimization, because the identity manager fetches
		// the data in parallel, which is a bit faster than fetching it in sequence.
		expand = identity.ExpandEverything
	}

	se, err := s.r.SessionPersister().GetSessionByToken(ctx, token, ExpandEverything, expand)
	if err != nil {
		if errors.Is(err, herodot.ErrNotFound) || errors.Is(err, sqlcon.ErrNoRows) {
			return nil, errors.WithStack(NewErrNoActiveSessionFound())
		}
		return nil, err
	}

	if !se.IsActive() {
		return nil, errors.WithStack(NewErrNoActiveSessionFound())
	}

	return se, nil
}

func (s *ManagerHTTP) PurgeFromRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
	ctx, span := s.r.Tracer(ctx).Tracer().Start(ctx, "sessions.ManagerHTTP.PurgeFromRequest")
	defer otelx.End(span, &err)

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

func (s *ManagerHTTP) DoesSessionSatisfy(r *http.Request, sess *Session, requestedAAL string, opts ...ManagerOptions) (err error) {
	ctx, span := s.r.Tracer(r.Context()).Tracer().Start(r.Context(), "sessions.ManagerHTTP.DoesSessionSatisfy")
	defer otelx.End(span, &err)

	managerOpts := &options{}

	for _, o := range opts {
		o(managerOpts)
	}

	sess.SetAuthenticatorAssuranceLevel()
	switch requestedAAL {
	case string(identity.AuthenticatorAssuranceLevel1):
		if sess.AuthenticatorAssuranceLevel >= identity.AuthenticatorAssuranceLevel1 {
			return nil
		}
	case config.HighestAvailableAAL:
		i := sess.Identity
		if i == nil {
			i, err = s.r.IdentityPool().GetIdentity(ctx, sess.IdentityID, identity.ExpandCredentials)
			if err != nil {
				return err
			}
			sess.Identity = i
		} else if len(i.Credentials) == 0 {
			// If credentials are not expanded, we load them here.
			if err := s.r.PrivilegedIdentityPool().HydrateIdentityAssociations(ctx, i, identity.ExpandCredentials); err != nil {
				return err
			}
		}

		available := identity.NoAuthenticatorAssuranceLevel
		if firstCount, err := s.r.IdentityManager().CountActiveFirstFactorCredentials(ctx, i); err != nil {
			return err
		} else if firstCount > 0 {
			available = identity.AuthenticatorAssuranceLevel1
		}

		if secondCount, err := s.r.IdentityManager().CountActiveMultiFactorCredentials(ctx, i); err != nil {
			return err
		} else if secondCount > 0 {
			available = identity.AuthenticatorAssuranceLevel2
		}

		if sess.AuthenticatorAssuranceLevel >= available {
			return nil
		}

		loginURL := urlx.CopyWithQuery(urlx.AppendPaths(s.r.Config().SelfPublicURL(ctx), "/self-service/login/browser"), url.Values{"aal": {"aal2"}})

		// return to the requestURL if it was set
		if managerOpts.requestURL != "" {
			loginURL = urlx.CopyWithQuery(loginURL, url.Values{"return_to": {managerOpts.requestURL}})
		}

		return NewErrAALNotSatisfied(loginURL.String())
	}

	return errors.Errorf("requested unknown aal: %s", requestedAAL)
}

func (s *ManagerHTTP) SessionAddAuthenticationMethods(ctx context.Context, sid uuid.UUID, ams ...AuthenticationMethod) (err error) {
	ctx, span := s.r.Tracer(ctx).Tracer().Start(ctx, "sessions.ManagerHTTP.SessionAddAuthenticationMethods")
	defer otelx.End(span, &err)

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

func (s *ManagerHTTP) MaybeRedirectAPICodeFlow(w http.ResponseWriter, r *http.Request, f flow.Flow, sessionID uuid.UUID, uiNode node.UiNodeGroup) (handled bool, err error) {
	ctx, span := s.r.Tracer(r.Context()).Tracer().Start(r.Context(), "sessions.ManagerHTTP.MaybeRedirectAPICodeFlow")
	defer otelx.End(span, &err)

	if uiNode != node.OpenIDConnectGroup {
		return false, nil
	}

	code, ok, _ := s.r.SessionTokenExchangePersister().CodeForFlow(ctx, f.GetID())
	if !ok {
		return false, nil
	}

	returnTo := s.r.Config().SelfServiceBrowserDefaultReturnTo(ctx)
	if redirecter, ok := f.(flow.FlowWithRedirect); ok {
		r, err := x.SecureRedirectTo(r, returnTo, redirecter.SecureRedirectToOpts(ctx, s.r)...)
		if err == nil {
			returnTo = r
		}
	}

	if err = s.r.SessionTokenExchangePersister().UpdateSessionOnExchanger(r.Context(), f.GetID(), sessionID); err != nil {
		return false, errors.WithStack(err)
	}

	q := returnTo.Query()
	q.Set("code", code.ReturnToCode)
	returnTo.RawQuery = q.Encode()
	http.Redirect(w, r, returnTo.String(), http.StatusSeeOther)

	return true, nil
}
