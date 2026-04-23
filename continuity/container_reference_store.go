// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package continuity

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"

	"github.com/ory/kratos/x"
)

// ContainerReferenceStore abstracts where a continuity container's ID is
// stored between Pause and Continue/Abort calls.
type ContainerReferenceStore interface {
	Store(ctx context.Context, w http.ResponseWriter, r *http.Request, name string, id uuid.UUID) error
	Retrieve(ctx context.Context, w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, error)
	Clear(ctx context.Context, w http.ResponseWriter, r *http.Request, name string) error
}

// CookieReferenceStore stores the container ID in an HTTP cookie.
type CookieReferenceStore struct {
	cs sessions.StoreExact
}

func NewCookieReferenceStore(cs sessions.StoreExact) *CookieReferenceStore {
	return &CookieReferenceStore{cs: cs}
}

func (s *CookieReferenceStore) Store(_ context.Context, w http.ResponseWriter, r *http.Request, name string, id uuid.UUID) error {
	return x.SessionPersistValues(w, r, s.cs, CookieName, map[string]any{
		name: id.String(),
	})
}

func (s *CookieReferenceStore) Retrieve(_ context.Context, w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, error) {
	str, err := x.SessionGetString(r, s.cs, CookieName, name)
	if err != nil {
		_ = x.SessionUnsetKey(w, r, s.cs, CookieName, name)
		return uuid.Nil, errors.WithStack(ErrNotResumable().WithDebugf("%+v", err))
	}

	sid, err := uuid.FromString(str)
	if err != nil {
		_ = x.SessionUnsetKey(w, r, s.cs, CookieName, name)
		return uuid.Nil, errors.WithStack(ErrNotResumable().WithDebug("session id is not a valid uuid"))
	}

	return sid, nil
}

func (s *CookieReferenceStore) Clear(_ context.Context, w http.ResponseWriter, r *http.Request, name string) error {
	return x.SessionUnsetKey(w, r, s.cs, CookieName, name)
}
