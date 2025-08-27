// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/kratos/x/redir"

	"go.opentelemetry.io/otel/attribute"

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

var ErrNoAALAvailable = herodot.ErrForbidden.WithReasonf("Unable to detect available authentication methods. Perform account recovery or contact support.")

type (
	managerHTTPDependencies interface {
		config.Provider
		identity.PoolProvider
		identity.PrivilegedPoolProvider
		identity.ManagementProvider
		x.CookieProvider
		x.LoggingProvider
		nosurfx.CSRFProvider
		x.TracingProvider
		x.TransactionPersistenceProvider
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
	upsertAAL  bool
}

type ManagerOptions func(*options)

// WithRequestURL passes along query parameters from the requestURL to the new URL (if any exist)
func WithRequestURL(requestURL string) ManagerOptions {
	return func(opts *options) {
		opts.requestURL = requestURL
	}
}

// UpsertAAL will update the available AAL of the identity if it was previoulsy unset. This is used to migrate
// identities from older versions of Ory Kratos.
func UpsertAAL(opts *options) {
	opts.upsertAAL = true
}

func (s *ManagerHTTP) UpsertAndIssueCookie(ctx context.Context, w http.ResponseWriter, r *http.Request, ss *Session) (err error) {
	ctx, span := s.r.Tracer(ctx).Tracer().Start(ctx, "sessions.ManagerHTTP.UpsertAndIssueCookie")
	defer otelx.End(span, &err)

	if err := s.r.SessionPersister().UpsertSession(ctx, ss); err != nil {
		return err
	}

	if err := s.IssueCookie(ctx, w, r, ss); err != nil {
		return err
	}

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
	ctx, span := s.r.Tracer(r.Context()).Tracer().Start(r.Context(), "sessions.ManagerHTTP.extractToken")
	defer span.End()

	if token := r.Header.Get("X-Session-Token"); len(token) > 0 {
		return token
	}

	cookie, err := s.getCookie(r.WithContext(ctx))
	if err != nil {
		token, _ := bearerTokenFromRequest(r.WithContext(ctx))
		return token
	}

	token, ok := cookie.Values["session_token"].(string)
	if ok {
		return token
	}

	token, _ = bearerTokenFromRequest(r.WithContext(ctx))
	return token
}

func (s *ManagerHTTP) FetchFromRequestContext(ctx context.Context, r *http.Request) (_ *Session, err error) {
	ctx, span := s.r.Tracer(ctx).Tracer().Start(ctx, "sessions.ManagerHTTP.FetchFromRequestContext")
	otelx.End(span, &err)

	if sess, ok := ctx.Value(sessionInContextKey).(*Session); ok {
		return sess, nil
	}

	return s.FetchFromRequest(ctx, r)
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

	se, err := s.r.SessionPersister().GetSessionByToken(ctx, token,
		// Don't change this unless you want bad performance down the line (because we constantly are unsure if we have the full data fetched or not).
		ExpandEverything, identity.ExpandEverything)
	if err != nil {
		if errors.Is(err, herodot.ErrNotFound) || errors.Is(err, sqlcon.ErrNoRows) {
			return nil, errors.WithStack(NewErrNoActiveSessionFound())
		}
		return nil, err
	}

	trace.SpanFromContext(ctx).AddEvent(events.NewSessionChecked(ctx, se.ID, se.IdentityID))

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

func (s *ManagerHTTP) DoesSessionSatisfy(ctx context.Context, sess *Session, requestedAAL string, opts ...ManagerOptions) (err error) {
	ctx, span := s.r.Tracer(ctx).Tracer().Start(ctx, "sessions.ManagerHTTP.DoesSessionSatisfy")
	defer otelx.End(span, &err)

	sess.SetAuthenticatorAssuranceLevel()

	// If we already have AAL2 there is no need to check further because it is the highest AAL.
	if sess.AuthenticatorAssuranceLevel == identity.AuthenticatorAssuranceLevel2 {
		return nil
	}

	managerOpts := &options{}
	for _, o := range opts {
		o(managerOpts)
	}

	loginURL := urlx.AppendPaths(s.r.Config().SelfPublicURL(ctx), "/self-service/login/browser")
	query := url.Values{
		"aal": {"aal2"},
	}

	// return to the requestURL if it was set
	if managerOpts.requestURL != "" {
		query.Set("return_to", managerOpts.requestURL)
	}

	loginURL.RawQuery = query.Encode()

	switch requestedAAL {
	case string(identity.AuthenticatorAssuranceLevel1):
		if sess.AuthenticatorAssuranceLevel >= identity.AuthenticatorAssuranceLevel1 {
			return nil
		}
		return NewErrAALNotSatisfied(loginURL.String())
	case config.HighestAvailableAAL:
		if sess.AuthenticatorAssuranceLevel == identity.AuthenticatorAssuranceLevel2 {
			// The session has AAL2, nothing to check.
			return nil
		}

		// The session is AAL1, we asked for `highest_available` AAL, so the only thing we can do
		// is actually check what authentication methods the identity has.
		if sess.Identity == nil {
			// This is nil if the session did not expand the identity field.
			sess.Identity, err = s.r.IdentityPool().GetIdentity(ctx, sess.IdentityID, identity.ExpandNothing)
			if err != nil {
				return err
			}
		}

		if aal, ok := sess.Identity.InternalAvailableAAL.ToAAL(); ok && aal == identity.AuthenticatorAssuranceLevel2 {
			// Identity gives us AAL2, but the session is still AAL1. We need to upgrade the session.
			return NewErrAALNotSatisfied(loginURL.String())
		}

		// Identity AAL is not 2, we refresh:

		// The identity was apparently fetched without credentials. Let's hydrate them.
		if len(sess.Identity.Credentials) == 0 {
			if err := s.r.PrivilegedIdentityPool().HydrateIdentityAssociations(ctx, sess.Identity, identity.ExpandCredentials); err != nil {
				return err
			}
		}

		// Great, now we determine the identity's available AAL
		if err := sess.Identity.SetAvailableAAL(ctx, s.r.IdentityManager()); err != nil {
			return err
		}

		// We override the result with our newly computed values
		available, valid := sess.Identity.InternalAvailableAAL.ToAAL()
		if !valid {
			// Unlikely to happen because SetAvailableAAL will either return an error, or a valid value - but not no error and an invalid value.
			return errors.WithStack(x.PseudoPanic.WithReasonf("Unable to determine available authentication methods for session: %s", sess.ID))
		}

		switch available {
		case identity.NoAuthenticatorAssuranceLevel:
			// The identity has AAL0, the session has AAL1, we're good.
			return nil
		case identity.AuthenticatorAssuranceLevel1:
			// The identity has AAL1, the session has AAL1, we're good.
			return nil
		case identity.AuthenticatorAssuranceLevel2:
			// The identity has AAL2, the session has AAL1, we need to upgrade the session.

			// Since we ended up here, it also means that `sess.Identity.InternalAvailableAAL` was `aal1` and is now `aal2`.
			// Let's update the database.
			if managerOpts.upsertAAL {
				if err := s.r.PrivilegedIdentityPool().UpdateIdentityColumns(ctx, sess.Identity, "available_aal"); err != nil {
					return err
				}
			}
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
		sess.CompletedLoginForMethod(m)
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
		r, err := redir.SecureRedirectTo(r, returnTo, redirecter.SecureRedirectToOpts(ctx, s.r)...)
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

func (s *ManagerHTTP) ActivateSession(r *http.Request, session *Session, i *identity.Identity, authenticatedAt time.Time) (err error) {
	ctx, span := s.r.Tracer(r.Context()).Tracer().Start(r.Context(), "sessions.ManagerHTTP.ActivateSession", trace.WithAttributes(
		attribute.String("session.id", session.ID.String()),
		attribute.String("identity.id", session.ID.String()),
		attribute.String("authenticated_at", session.ID.String()),
	))
	defer otelx.End(span, &err)

	if i == nil {
		return errors.WithStack(x.PseudoPanic.WithReasonf("Identity must not be nil when activating a session."))
	}

	if !i.IsActive() {
		return errors.WithStack(ErrIdentityDisabled.WithDetail("identity_id", i.ID))
	}

	if err := s.r.IdentityManager().RefreshAvailableAAL(ctx, i); err != nil {
		return err
	}

	session.Identity = i
	session.IdentityID = i.ID

	session.Active = true
	session.IssuedAt = authenticatedAt
	session.ExpiresAt = authenticatedAt.Add(s.r.Config().SessionLifespan(ctx))
	session.AuthenticatedAt = authenticatedAt

	session.SetSessionDeviceInformation(r.WithContext(ctx))
	session.SetAuthenticatorAssuranceLevel()

	span.SetAttributes(
		attribute.String("identity.available_aal", session.Identity.InternalAvailableAAL.String),
	)

	return nil
}
