package session

import (
	"context"
	"net/http"

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
	CreateToRequest(context.Context, http.ResponseWriter, *http.Request, *Session) error

	// SaveToRequest creates an HTTP session using cookies.
	SaveToRequest(context.Context, http.ResponseWriter, *http.Request, *Session) error

	// FetchFromRequest creates an HTTP session using cookies.
	FetchFromRequest(context.Context, *http.Request) (*Session, error)

	// PurgeFromRequest removes an HTTP session.
	PurgeFromRequest(context.Context, http.ResponseWriter, *http.Request) error
}

type ManagementProvider interface {
	SessionManager() Manager
}
