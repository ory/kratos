package session

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/hive/identity"

	"github.com/ory/herodot"
)

type ManagerMemory struct {
	sessions   map[string]Session
	c          Configuration
	cookieName string
	r          Registry
}

func NewManagerMemory(c Configuration, r Registry) *ManagerMemory {
	return &ManagerMemory{
		sessions:   make(map[string]Session),
		c:          c,
		r:          r,
		cookieName: DefaultSessionCookieName,
	}
}

func (s *ManagerMemory) Get(ctx context.Context, sid string) (*Session, error) {
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

func (s *ManagerMemory) CreateToRequest(ctx context.Context, i *identity.Identity, w http.ResponseWriter, r *http.Request) (*Session, error) {
	p := NewSession(i, r, s.c)
	if err := s.Create(ctx, p); err != nil {
		return nil, err
	}

	if err := s.SaveToRequest(ctx, p, w, r); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *ManagerMemory) Create(ctx context.Context, session *Session) error {
	insert := *session
	insert.Identity = insert.Identity.CopyWithoutCredentials()
	s.sessions[session.SID] = insert
	return nil
}

func (s *ManagerMemory) Delete(ctx context.Context, sid string) error {
	delete(s.sessions, sid)
	return nil
}

func (s *ManagerMemory) SaveToRequest(ctx context.Context, session *Session, w http.ResponseWriter, r *http.Request) error {
	cookie, _ := s.r.CookieManager().Get(r, s.cookieName)
	cookie.Values["sid"] = session.SID
	if err := cookie.Save(r, w); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *ManagerMemory) FetchFromRequest(ctx context.Context, r *http.Request) (*Session, error) {
	cookie, err := s.r.CookieManager().Get(r, s.cookieName)
	if err != nil {
		return nil, errors.WithStack(ErrNoActiveSessionFound.WithDebug(err.Error()))
	}

	sid, ok := cookie.Values["sid"].(string)
	if !ok {
		return nil, errors.WithStack(ErrNoActiveSessionFound)
	}

	se, err := s.Get(ctx, sid)
	if err != nil && err.Error() == herodot.ErrNotFound.Error() {
		return nil, errors.WithStack(ErrNoActiveSessionFound)
	} else if err != nil {
		return nil, err
	}

	se.Identity = se.Identity.CopyWithoutCredentials()

	return se, nil
}

func (s *ManagerMemory) PurgeFromRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	cookie, _ := s.r.CookieManager().Get(r, s.cookieName)
	cookie.Options.MaxAge = -1
	if err := cookie.Save(r, w); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
