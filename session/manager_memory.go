package session

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
)

var _ Manager = new(ManagerMemory)

type ManagerMemory struct {
	*ManagerHTTP
	sync.RWMutex
	sessions map[string]Session
}

func NewManagerMemory(c Configuration, r Registry) *ManagerMemory {
	m := &ManagerMemory{sessions: make(map[string]Session)}
	m.ManagerHTTP = NewManagerHTTP(c, r, m)
	return m
}

func (s *ManagerMemory) Get(ctx context.Context, sid string) (*Session, error) {
	s.RLock()
	defer s.RUnlock()
	if r, ok := s.sessions[sid]; ok {
		i, err := s.r.IdentityPool().Get(ctx, r.Identity.ID)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		r.Identity = i
		return &r, nil
	}

	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Unable to find session with id: %s", sid))
}

func (s *ManagerMemory) Create(ctx context.Context, session *Session) error {
	s.Lock()
	defer s.Unlock()
	insert := *session
	insert.Identity = insert.Identity.CopyWithoutCredentials()
	s.sessions[session.SID] = insert
	return nil
}

func (s *ManagerMemory) Delete(ctx context.Context, sid string) error {
	s.Lock()
	defer s.Unlock()
	delete(s.sessions, sid)
	return nil
}
