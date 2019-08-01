package errorx

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ory/x/urlx"

	"github.com/pkg/errors"

	"github.com/ory/hive/driver/configuration"

	"github.com/google/uuid"

	"github.com/ory/herodot"
)

type container struct {
	errs   []error
	read   bool
	readAt time.Time
}

var _ Manager = new(MemoryManager)

type MemoryManager struct {
	sync.RWMutex
	containers map[string]container
	l          logrus.FieldLogger
	w          herodot.Writer
	c          configuration.Provider
}

func NewMemoryManager(
	l logrus.FieldLogger,
	w herodot.Writer,
	c configuration.Provider,
) Manager {
	return &MemoryManager{
		containers: make(map[string]container),
		l:          l,
		w:          w,
		c:          c,
	}
}

func (m *MemoryManager) ForwardError(w http.ResponseWriter, r *http.Request, errs ...error) {
	for _, err := range errs {
		herodot.DefaultErrorLogger(m.l, err).Errorf("An error occurred and is being forwarded to the error user interface.")
	}

	id, emerr := m.Add(errs...)
	if emerr != nil {
		m.w.WriteError(w, r, emerr)
		return
	}
	q := url.Values{}
	q.Set("error", id)

	to := urlx.CopyWithQuery(m.c.ErrorURL(), q).String()
	http.Redirect(w, r, to, http.StatusFound)
}

func (m *MemoryManager) Add(errs ...error) (string, error) {
	es := make([]error, len(errs))
	for k, e := range errs {
		if e == nil {
			return "", herodot.ErrInternalServerError.WithDebug("A nil error was passed to the error manager which is most likely a code bug.")
		}
		es[k] = errors.Cause(e)
	}

	id := uuid.New().String()

	m.Lock()
	m.containers[id] = container{
		errs: es,
	}
	m.Unlock()

	return id, nil
}

func (m *MemoryManager) Read(id string) ([]error, error) {
	m.RLock()
	c, ok := m.containers[id]
	m.RUnlock()
	if !ok {
		return nil, herodot.ErrNotFound.WithReasonf("Unable to find error with id: %s", id)
	}

	c.read = true
	c.readAt = time.Now()

	m.Lock()
	m.containers[id] = c
	m.Unlock()

	return c.errs, nil
}

func (m *MemoryManager) Clear(olderThan time.Duration, force bool) error {
	m.Lock()
	defer m.Unlock()
	for k, c := range m.containers {
		if (c.read || force) && c.readAt.Before(time.Now().Add(-olderThan)) {
			delete(m.containers, k)
		}
	}

	return nil
}
