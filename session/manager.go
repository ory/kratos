package session

import (
	"context"
	"net/http"

	"github.com/ory/hive/identity"

	"github.com/ory/herodot"
)

// DefaultSessionCookieName returns the default cookie name for the hive session.
const DefaultSessionCookieName = "hive_session_manager"

var (
	ErrNoActiveSessionFound = herodot.ErrUnauthorized.WithReason("No active session was found in this request.")
)

// Manager handles identity sessions.
type Manager interface {
	// Get retrieves a session from the store.
	Get(ctx context.Context, sid string) (*Session, error)

	// Create adds a session to the store.
	Create(ctx context.Context, s *Session) error

	// Delete removes a session from the store
	Delete(ctx context.Context, sid string) error

	CreateToRequest(context.Context, *identity.Identity, http.ResponseWriter, *http.Request) (*Session, error)

	// SaveToRequest creates an HTTP session using cookies.
	SaveToRequest(context.Context, *Session, http.ResponseWriter, *http.Request) error

	// FetchFromRequest creates an HTTP session using cookies.
	FetchFromRequest(context.Context, *http.Request) (*Session, error)

	// PurgeFromRequest removes an HTTP session.
	PurgeFromRequest(context.Context, http.ResponseWriter, *http.Request) error
}

type ManagementProvider interface {
	SessionManager() Manager
}
