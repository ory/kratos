package errorx

import (
	"net/http"
	"time"
)

type Manager interface {
	// Add adds an error to the manager and returns a unique identifier or an error if insertion fails.
	Add(...error) (string, error)

	// Read returns an error by its unique identifier and marks the error as read. If an error occurs during retrieval
	// the second return parameter is an error.
	Read(id string) ([]error, error)

	// Clear clears read containers that are older than a certain amount of time. If force is set to true, unread
	// errors will be cleared as well.
	Clear(olderThan time.Duration, force bool) error

	ForwardError(http.ResponseWriter, *http.Request, ...error)
}

type ManagementProvider interface {
	ErrorManager() Manager
}
