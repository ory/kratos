package errorx

import (
	"bytes"
	"context"
	"encoding/json"
	stderr "errors"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/hive/schema"
	"github.com/ory/hive/x"
)

type (
	Manager interface {
		Store

		// ForwardError is a simple helper that saves all errors in the store and forwards the HTTP Request
		// to the error url, appending the error ID.
		ForwardError(ctx context.Context, rw http.ResponseWriter, r *http.Request, errs ...error)
	}

	ManagementProvider interface {
		// ErrorManager returns the errorx.Manager.
		ErrorManager() Manager
	}

	Store interface {
		// Add adds an error to the manager and returns a unique identifier or an error if insertion fails.
		Add(ctx context.Context, errs ...error) (string, error)

		// Read returns an error by its unique identifier and marks the error as read. If an error occurs during retrieval
		// the second return parameter is an error.
		Read(ctx context.Context, id string) ([]json.RawMessage, error)

		// Clear clears read containers that are older than a certain amount of time. If force is set to true, unread
		// errors will be cleared as well.
		Clear(ctx context.Context, olderThan time.Duration, force bool) error
	}

	baseManagerDependencies interface {
		x.LoggingProvider
		x.WriterProvider
	}

	baseManagerConfiguration interface {
		ErrorURL() *url.URL
	}

	BaseManager struct {
		Store
		d baseManagerDependencies
		c baseManagerConfiguration
	}
)

func NewBaseManager(d baseManagerDependencies, c baseManagerConfiguration, m Store) *BaseManager {
	return &BaseManager{d: d, c: c, Store: m}
}

func (m *BaseManager) ForwardError(ctx context.Context, w http.ResponseWriter, r *http.Request, errs ...error) {
	for _, err := range errs {
		herodot.DefaultErrorLogger(m.d.Logger(), err).Errorf("An error occurred and is being forwarded to the error user interface.")
	}

	id, emerr := m.Add(ctx, errs...)
	if emerr != nil {
		m.d.Writer().WriteError(w, r, emerr)
		return
	}
	q := url.Values{}
	q.Set("error", id)

	to := urlx.CopyWithQuery(m.c.ErrorURL(), q).String()
	http.Redirect(w, r, to, http.StatusFound)
}

func (m *BaseManager) encode(errs []error) (*bytes.Buffer, error) {
	es := make([]interface{}, len(errs))
	for k, e := range errs {
		e = errors.Cause(e)
		if u := stderr.Unwrap(e); u != nil {
			e = u
		}

		if e == nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithDebug("A nil error was passed to the error manager which is most likely a code bug."))
		}

		// Convert to a default error if the error type is unknown. Helps to properly
		// pass through system errors.
		switch e.(type) {
		case *herodot.DefaultError:
		case *schema.ResultErrors:
		case schema.ResultErrors:
		default:
			e = herodot.ToDefaultError(e, "")
		}

		es[k] = e
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(es); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to encode error messages.").WithDebug(err.Error()))
	}

	return &b, nil
}
