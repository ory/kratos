package errorx

import (
	"context"
	"net/http"
	"net/url"

	"github.com/justinas/nosurf"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/x"
)

type (
	managerDependencies interface {
		PersistenceProvider
		x.LoggingProvider
		x.WriterProvider
	}

	Manager struct {
		d    managerDependencies
		c    baseManagerConfiguration
		csrf x.CSRFToken
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
	return &Manager{d: d, c: c, csrf: nosurf.Token}
}

func (m *Manager) WithTokenGenerator(f func(r *http.Request) string) {
	m.csrf = f
}

// ForwardError is a simple helper that saves all errors in the store and forwards the HTTP Request
// to the error url, appending the error ID.
func (m *Manager) ForwardError(ctx context.Context, w http.ResponseWriter, r *http.Request, errs ...error) {
	for _, err := range errs {
		herodot.DefaultErrorLogger(m.d.Logger(), err).Errorf("An error occurred and is being forwarded to the error user interface.")
	}

	id, emerr := m.d.SelfServiceErrorPersister().Add(ctx, m.csrf(r), errs...)
	if emerr != nil {
		m.d.Writer().WriteError(w, r, emerr)
		return
	}
	q := url.Values{}
	q.Set("error", id.String())

	to := urlx.CopyWithQuery(m.c.ErrorURL(), q).String()
	http.Redirect(w, r, to, http.StatusFound)
}
