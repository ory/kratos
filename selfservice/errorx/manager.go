package errorx

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/x"
)

type (
	managerDependencies interface {
		PersistenceProvider
		x.LoggingProvider
		x.WriterProvider
		x.CSRFTokenGeneratorProvider
	}

	Manager struct {
		d managerDependencies
		c baseManagerConfiguration
	}

	ManagementProvider interface {
		// SelfServiceErrorManager returns the errorx.Manager.
		SelfServiceErrorManager() *Manager
	}

	baseManagerConfiguration interface {
		ErrorURL() *url.URL
	}
)

func NewManager(d managerDependencies, c baseManagerConfiguration) *Manager {
	return &Manager{d: d, c: c}
}

// Create is a simple helper that saves all errors in the store and returns the
// error url, appending the error ID.
func (m *Manager) Create(ctx context.Context, w http.ResponseWriter, r *http.Request, errs ...error) (string, error) {
	for _, err := range errs {
		herodot.DefaultErrorLogger(m.d.Logger(), err).Errorf("An error occurred and is being forwarded to the error user interface.")
	}

	id, emerr := m.d.SelfServiceErrorPersister().Add(ctx, m.d.GenerateCSRFToken(r), errs...)
	if emerr != nil {
		return "", emerr
	}
	q := url.Values{}
	q.Set("error", id.String())

	return urlx.CopyWithQuery(m.c.ErrorURL(), q).String(), nil
}

// Forward is a simple helper that saves all errors in the store and forwards the HTTP Request
// to the error url, appending the error ID.
func (m *Manager) Forward(ctx context.Context, w http.ResponseWriter, r *http.Request, errs ...error) {
	to, err := m.Create(ctx, w, r, errs...)
	if err != nil {
		// Everything failed. Resort to standard error output.
		m.d.Writer().WriteError(w, r, err)
		return
	}
	http.Redirect(w, r, to, http.StatusFound)
}
