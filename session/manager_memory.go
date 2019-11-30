package session

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
)

var _ Manager = new(ManagerMemory)

type ManagerMemory struct {
	*ManagerHTTP
	sync.RWMutex
	sessions map[uuid.UUID]Session
}

func NewManagerMemory(c Configuration, r Registry) *ManagerMemory {
	m := &ManagerMemory{sessions: make(map[uuid.UUID]Session)}
	// m.ManagerHTTP = NewManagerHTTP(c, r)
	return m
}

func (s *ManagerMemory) GetSession(ctx context.Context, sid uuid.UUID) (*Session, error) {
	s.RLock()
	defer s.RUnlock()
	if r, ok := s.sessions[sid]; ok {
		i, err := s.r.IdentityPool().GetIdentity(ctx, r.Identity.ID)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		r.Identity = i
		return &r, nil
	}

	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Unable to find session with id: %s", sid))
}

func (s *ManagerMemory) CreateSession(ctx context.Context, session *Session) error {
	s.Lock()
	defer s.Unlock()
	insert := *session
	insert.Identity = insert.Identity.CopyWithoutCredentials()
	s.sessions[session.ID] = insert
	return nil
}

func (s *ManagerMemory) DeleteSession(ctx context.Context, sid uuid.UUID) error {
	s.Lock()
	defer s.Unlock()
	delete(s.sessions, sid)
	return nil
}
