// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/kratos/identity"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x/swagger"

	"github.com/gofrs/uuid"

	"github.com/ory/herodot"
)

// ErrNoActiveSessionFound is returned when no active cookie session could be found in the request.
type ErrNoActiveSessionFound struct {
	*herodot.DefaultError `json:"error"`

	// True when the request had no credentials in it.
	CredentialsMissing bool
}

// NewErrNoActiveSessionFound creates a new ErrNoActiveSessionFound
func NewErrNoActiveSessionFound() *ErrNoActiveSessionFound {
	return &ErrNoActiveSessionFound{
		DefaultError: herodot.ErrUnauthorized().WithID(text.ErrIDNoActiveSession).WithError("request does not have a valid authentication session").WithReason("No active session was found in this request."),
	}
}

// NewErrNoCredentialsForSession creates a new NewErrNoCredentialsForSession
func NewErrNoCredentialsForSession() *ErrNoActiveSessionFound {
	e := NewErrNoActiveSessionFound()
	e.CredentialsMissing = true
	return e
}

func (e *ErrNoActiveSessionFound) EnhanceJSONError() interface{} {
	return e
}

// Is returned when an active session was found but the requested AAL is not satisfied.
//
// swagger:model errorAuthenticatorAssuranceLevelNotSatisfied
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type errorAuthenticatorAssuranceLevelNotSatisfied struct {
	Error swagger.GenericError `json:"error"`

	// Points to where to redirect the user to next.
	RedirectTo string `json:"redirect_browser_to"`
}

// ErrAALNotSatisfied is returned when an active session was found but the requested AAL is not satisfied.
type ErrAALNotSatisfied struct {
	*herodot.DefaultError `json:"error"`
	RedirectTo            string `json:"redirect_browser_to"`
}

func (e *ErrAALNotSatisfied) EnhanceJSONError() interface{} {
	return e
}

func (e *ErrAALNotSatisfied) PassReturnToAndLoginChallengeParameters(requestURL string) error {
	req, err := url.Parse(requestURL)
	if err != nil {
		return err
	}

	u, err := url.Parse(e.RedirectTo)
	if err != nil {
		return err
	}
	q := u.Query()

	hlc := req.Query().Get("login_challenge")
	if len(hlc) != 0 {
		q.Set("login_challenge", hlc)
	}

	returnTo := req.Query().Get("return_to")
	if len(returnTo) != 0 {
		q.Set("return_to", returnTo)
	}

	u.RawQuery = q.Encode()
	e.RedirectTo = u.String()

	return nil
}

// NewErrAALNotSatisfied creates a new ErrAALNotSatisfied.
func NewErrAALNotSatisfied(redirectTo string) *ErrAALNotSatisfied {
	return &ErrAALNotSatisfied{
		RedirectTo: redirectTo,
		DefaultError: &herodot.DefaultError{
			IDField:     text.ErrIDHigherAALRequired,
			StatusField: http.StatusText(http.StatusForbidden),
			ErrorField:  "Session does not fulfill the requested Authenticator Assurance Level",
			ReasonField: "An active session was found but it does not fulfill the requested Authenticator Assurance Level. Please verify yourself with a second factor to resolve this issue.",
			CodeField:   http.StatusForbidden,
			DetailsField: map[string]interface{}{
				"redirect_browser_to": redirectTo,
			},
		},
	}
}

// Manager handles identity sessions.
type Manager interface {
	// UpsertAndIssueCookie stores a session in the database and issues a cookie by calling IssueCookie.
	//
	// Also regenerates CSRF tokens due to assumed principal change.
	UpsertAndIssueCookie(context.Context, http.ResponseWriter, *http.Request, *Session) error

	// IssueCookie issues a cookie for the given session.
	//
	// Also regenerates CSRF tokens due to assumed principal change.
	IssueCookie(context.Context, http.ResponseWriter, *http.Request, *Session) error

	// RefreshCookie checks if the request uses an outdated cookie and refreshes the cookie if needed.
	RefreshCookie(context.Context, http.ResponseWriter, *http.Request, *Session) error

	// FetchFromRequest creates an HTTP session using cookies.
	//
	// sessionExpand controls which session associations (devices, identity) are loaded, and
	// identityExpand controls which identity relations are loaded when sessionExpand includes
	// ExpandSessionIdentity. Pass ExpandEverything/identity.ExpandEverything unless you are
	// certain the caller never reads an association: expanding the identity's credentials in
	// particular is by far the most expensive part of fetching a session, so hot paths that
	// never read Identity.Credentials (such as /sessions/whoami) can pass
	// identity.ExpandEverythingButCredentials, and callers that only need session columns
	// (ID, IdentityID, AAL, ...) can pass ExpandNothing/identity.ExpandNothing. AAL
	// enforcement stays intact: DoesSessionSatisfy loads the identity and hydrates the
	// credentials lazily when it has to.
	FetchFromRequest(ctx context.Context, r *http.Request, sessionExpand Expandables, identityExpand identity.Expandables) (*Session, error)

	// FetchFromRequestContext returns the session from the context or if that is unset, falls back to FetchFromRequest.
	FetchFromRequestContext(context.Context, *http.Request) (*Session, error)

	// SessionActiveForRequest reports whether the request carries an active session, without loading the identity.
	SessionActiveForRequest(context.Context, *http.Request) error

	// PurgeFromRequest removes an HTTP session.
	PurgeFromRequest(context.Context, http.ResponseWriter, *http.Request) error

	// DoesSessionSatisfy answers if a session is satisfying the AAL of a user.
	//
	// The matcher value can be one of:
	//
	// - `highest_available`: If set requires the user to upgrade their session to the highest available AAL for that user.
	// - `aal1`: Requires the user to have authenticated with at least one authentication factor.
	//
	// This method is implemented in such a way, that if a second factor is found for the user, it is always assumed
	// that the user is able to authenticate with it. This means that if a user has a second factor, the user is always
	// asked to authenticate with it if `highest_available` is set and the session's AAL is `aal1`.
	DoesSessionSatisfy(ctx context.Context, sess *Session, matcher string, opts ...ManagerOptions) error

	// SessionAddAuthenticationMethods adds one or more authentication method to the session.
	SessionAddAuthenticationMethods(ctx context.Context, sid uuid.UUID, methods ...AuthenticationMethod) error

	// MaybeRedirectAPICodeFlow for API+Code flows redirects the user to the return_to URL and adds the code query parameter.
	// `handled` is true if the request a redirect was written, false otherwise.
	MaybeRedirectAPICodeFlow(w http.ResponseWriter, r *http.Request, f flow.Flow, sessionID uuid.UUID, uiNode node.UiNodeGroup) (handled bool, err error)

	// ActivateSession activates a session.
	//
	// This method is used to activate a session after a user authenticated with a first or second factor. It sets
	// all computed values (e.g. authenticator assurance level) and updates the session object but does not store
	// the session in the database or on the client device.
	ActivateSession(r *http.Request, session *Session, i *identity.Identity, authenticatedAt time.Time) error

	// IsPrivileged checks if a session can be considered privileged.
	// https://ory.com/docs/kratos/session-management/session-lifespan#privileged-sessions
	IsPrivileged(ctx context.Context, session *Session) bool
}

type ManagementProvider interface {
	SessionManager() Manager
}
