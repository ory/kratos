package errorx

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/google/uuid"

	"github.com/ory/herodot"
)

var _ Manager = new(ManagerMemory)

type (
	containerMemory struct {
		errs   []byte
		read   bool
		readAt time.Time
	}

	ManagerMemory struct {
		sync.RWMutex
		containers map[string]containerMemory
		*BaseManager
	}
)

func NewManagerMemory(
	d baseManagerDependencies,
	c baseManagerConfiguration,
) *ManagerMemory {
	m := &ManagerMemory{containers: make(map[string]containerMemory)}
	m.BaseManager = NewBaseManager(d, c, m)
	return m
}

func (m *ManagerMemory) Add(ctx context.Context, errs ...error) (string, error) {
	b, err := m.encode(errs)
	if err != nil {
		return "", err
	}

	id := uuid.New().String()

	m.Lock()
	m.containers[id] = containerMemory{
		errs: b.Bytes(),
	}
	m.Unlock()

	return id, nil
}

func (m *ManagerMemory) Read(ctx context.Context, id string) ([]json.RawMessage, error) {
	m.RLock()
	c, ok := m.containers[id]
	m.RUnlock()
	if !ok {
		return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Unable to find error with id: %s", id))
	}

	c.read = true
	c.readAt = time.Now()

	m.Lock()
	m.containers[id] = c
	m.Unlock()

	var errs []json.RawMessage
	if err := json.NewDecoder(bytes.NewReader(c.errs)).Decode(&errs); err != nil {
		return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Unable to decode errors.").WithDebug(err.Error()))
	}

	return errs, nil
}

func (m *ManagerMemory) Clear(ctx context.Context, olderThan time.Duration, force bool) error {
	m.Lock()
	defer m.Unlock()
	for k, c := range m.containers {
		if (c.read || force) && c.readAt.Before(time.Now().Add(-olderThan)) {
			delete(m.containers, k)
		}
	}

	return nil
}
