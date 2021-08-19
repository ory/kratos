package session

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/identity"

	"github.com/ory/herodot"
)

var (
	// ErrNoActiveSessionFound is returned when no active cookie session could be found in the request.
	ErrNoActiveSessionFound = herodot.ErrUnauthorized.WithError("request does not have a valid authentication session").WithReason("No active session was found in this request.")

	// ErrAALNotSatisfied is returned when an active session was found but the requested AAL is not satisfied.
	ErrAALNotSatisfied = herodot.ErrForbidden.WithError("session does not fulfill the requested AAL").WithReason("An active session was found but it does not fulfill the requested Authenticator Assurance Level. You might be able to resolve this by signing in with 2FA.")
)

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

	// FetchFromRequest creates an HTTP session using cookies.
	FetchFromRequest(context.Context, *http.Request) (*Session, error)

	// PurgeFromRequest removes an HTTP session.
	PurgeFromRequest(context.Context, http.ResponseWriter, *http.Request) error

	// DoesSessionSatisfy answers if a session is satisfying the AAL.
	DoesSessionSatisfy(ctx context.Context, sess *Session, requestedAAL string) error

	// SessionAddAuthenticationMethod adds one or more authentication method to the session.
	SessionAddAuthenticationMethod(ctx context.Context, sid uuid.UUID, method ...identity.CredentialsType) error
}

type ManagementProvider interface {
	SessionManager() Manager
}
