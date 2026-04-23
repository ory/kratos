// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package continuity

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/kratos/selfservice/flow"
)

const InternalContextKey = "continuity_container_id"

// InternalContextReferenceStore stores the container ID in a flow's
// InternalContext. Used for native/API flows where cookies are not available.
// This store only modifies the in-memory InternalContext — the caller is
// responsible for persisting the flow to the database.
type InternalContextReferenceStore struct {
	ic flow.InternalContexter
}

func NewInternalContextReferenceStore(ic flow.InternalContexter) *InternalContextReferenceStore {
	return &InternalContextReferenceStore{ic: ic}
}

func (s *InternalContextReferenceStore) Store(_ context.Context, _ http.ResponseWriter, _ *http.Request, _ string, id uuid.UUID) error {
	s.ic.EnsureInternalContext()
	raw, err := sjson.SetBytes(s.ic.GetInternalContext(), InternalContextKey, id.String())
	if err != nil {
		return errors.WithStack(err)
	}
	s.ic.SetInternalContext(raw)
	return nil
}

func (s *InternalContextReferenceStore) Retrieve(_ context.Context, _ http.ResponseWriter, _ *http.Request, _ string) (uuid.UUID, error) {
	cidStr := gjson.GetBytes(s.ic.GetInternalContext(), InternalContextKey).String()
	if cidStr == "" {
		return uuid.Nil, errors.WithStack(ErrNotResumable().WithDebug("no continuity container ID in flow internal context"))
	}
	cid, err := uuid.FromString(cidStr)
	if err != nil {
		return uuid.Nil, errors.WithStack(herodot.ErrBadRequest().WithReasonf("Invalid continuity container ID in flow context."))
	}
	return cid, nil
}

// Clear is a no-op: InternalContext entries are cleaned up when the flow
// expires, so there is no separate reference to remove.
func (s *InternalContextReferenceStore) Clear(_ context.Context, _ http.ResponseWriter, _ *http.Request, _ string) error {
	return nil
}
