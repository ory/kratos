package session

import (
	"net/http"

	"github.com/ory/hive-cloud/hive/identity"

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
	Get(sid string) (*Session, error)

	// Create adds a session to the store.
	Create(*Session) error

	// Delete removes a session from the store
	Delete(sid string) error

	CreateToRequest(*identity.Identity, http.ResponseWriter, *http.Request) (*Session, error)

	// SaveToRequest creates an HTTP session using cookies.
	SaveToRequest(*Session, http.ResponseWriter, *http.Request) error

	// FetchFromRequest creates an HTTP session using cookies.
	FetchFromRequest(*http.Request) (*Session, error)

	// PurgeFromRequest removes an HTTP session.
	PurgeFromRequest(http.ResponseWriter, *http.Request) error
}

type ManagementProvider interface {
	SessionManager() Manager
}
