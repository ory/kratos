// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
)

const (
	PendingTraitsChangeStatusPending   = "pending"
	PendingTraitsChangeStatusCompleted = "completed"
)

// PendingTraitsChange stores a deferred traits update that is waiting
// for the new address to be verified before being applied to the identity.
//
// swagger:ignore
type PendingTraitsChange struct {
	// ID is the unique identifier of this pending change.
	ID uuid.UUID `json:"id" db:"id"`

	// IdentityID is the identity this change belongs to.
	IdentityID uuid.UUID `json:"identity_id" db:"identity_id"`

	// NID is the network ID (multi-tenant discriminator).
	NID uuid.UUID `json:"-" db:"nid"`

	// NewAddressValue is the new address value being verified (e.g., "new@example.com").
	NewAddressValue string `json:"new_address_value" db:"new_address_value"`

	// NewAddressVia is the delivery channel for the new address ("email" or "sms").
	NewAddressVia string `json:"new_address_via" db:"new_address_via"`

	// OriginalTraitsHash is the SHA-256 hash of the identity's traits at the time
	// the pending change was created. Used to detect concurrent modifications.
	OriginalTraitsHash string `json:"original_traits_hash" db:"original_traits_hash"`

	// ProposedTraits is the full traits JSON snapshot to apply on successful verification.
	ProposedTraits json.RawMessage `json:"proposed_traits" db:"proposed_traits"`

	// VerificationFlowID links this pending change to the verification flow that must complete.
	VerificationFlowID uuid.UUID `json:"verification_flow_id" db:"verification_flow_id"`

	// Status is the current state: "pending" or "completed".
	Status string `json:"status" db:"status"`

	// SessionID is the session that created this pending change. Used at apply
	// time to verify the session is still active and match it to the current NID.
	// Nullable: ON DELETE SET NULL in the schema — a NULL value means the
	// session was revoked or deleted and the pending change must be rejected.
	SessionID uuid.NullUUID `json:"session_id,omitempty" db:"session_id"`

	// OriginSettingsFlowID is the settings flow that produced this pending
	// change. Used at apply time to load the real flow for the
	// settings-post-persist webhook chain. Non-nullable in practice: the
	// FK is declared ON DELETE CASCADE, so a missing flow deletes the PTC.
	// A NULL value here means the FK was never set by a malformed writer
	// and the pending change must be rejected.
	OriginSettingsFlowID uuid.NullUUID `json:"origin_settings_flow_id,omitempty" db:"origin_settings_flow_id"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (PendingTraitsChange) TableName() string {
	return "identity_pending_traits_changes"
}

func (p *PendingTraitsChange) ToPersistable() (*VerifiableAddress, bool) {
	return nil, false
}

func (p *PendingTraitsChange) Address() string {
	return p.NewAddressValue
}

func (p *PendingTraitsChange) DeliveryVia() string {
	return p.NewAddressVia
}

// HashTraits returns a hex-encoded SHA-256 hash of the JSON-compacted traits.
// Used to detect concurrent identity modifications between pending change
// creation and application.
func HashTraits(traits json.RawMessage) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, traits); err != nil {
		h := sha256.Sum256(traits)
		return hex.EncodeToString(h[:])
	}
	h := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(h[:])
}

// PendingTraitsChangePersister handles CRUD for pending traits changes.
type PendingTraitsChangePersister interface {
	CreatePendingTraitsChange(ctx context.Context, p *PendingTraitsChange) error
	GetPendingTraitsChangeByVerificationFlow(ctx context.Context, flowID uuid.UUID) (*PendingTraitsChange, error)
	DeletePendingTraitsChangesByIdentity(ctx context.Context, identityID uuid.UUID) error
	UpdatePendingTraitsChange(ctx context.Context, p *PendingTraitsChange) error
}

// PendingTraitsChangePersistenceProvider provides access to the persister.
type PendingTraitsChangePersistenceProvider interface {
	PendingTraitsChangePersister() PendingTraitsChangePersister
}
