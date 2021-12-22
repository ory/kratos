package link

import (
	"context"
	"time"

	"github.com/ory/kratos/selfservice/flow"

	"github.com/ory/kratos/corp"

	"github.com/gofrs/uuid"
	errors "github.com/pkg/errors"

	"github.com/ory/x/randx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/x"
)

type RecoveryToken struct {
	// ID represents the tokens's unique ID.
	//
	// required: true
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// Token represents the recovery token. It can not be longer than 64 chars!
	Token string `json:"-" db:"token"`

	// RecoveryAddress links this token to a recovery address.
	// required: true
	RecoveryAddress *identity.RecoveryAddress `json:"recovery_address" belongs_to:"identity_recovery_addresses" fk_id:"RecoveryAddressID"`

	// ExpiresAt is the time (UTC) when the token expires.
	// required: true
	ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

	// IssuedAt is the time (UTC) when the token was issued.
	// required: true
	IssuedAt time.Time `json:"issued_at" faker:"time_type" db:"issued_at"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
	// RecoveryAddressID is a helper struct field for gobuffalo.pop.
	RecoveryAddressID *uuid.UUID `json:"-" faker:"-" db:"identity_recovery_address_id"`
	// FlowID is a helper struct field for gobuffalo.pop.
	FlowID     uuid.NullUUID `json:"-" faker:"-" db:"selfservice_recovery_flow_id"`
	NID        uuid.UUID     `json:"-"  faker:"-" db:"nid"`
	IdentityID uuid.UUID     `json:"identity_id"  faker:"-" db:"identity_id"`
}

func (RecoveryToken) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "identity_recovery_tokens")
}

func NewSelfServiceRecoveryToken(address *identity.RecoveryAddress, f *recovery.Flow, expiresIn time.Duration) *RecoveryToken {
	now := time.Now().UTC()
	var identityID = uuid.UUID{}
	var recoveryAddressID = uuid.UUID{}
	if address != nil {
		identityID = address.IdentityID
		recoveryAddressID = address.ID
	}
	return &RecoveryToken{
		ID:                x.NewUUID(),
		Token:             randx.MustString(32, randx.AlphaNum),
		RecoveryAddress:   address,
		ExpiresAt:         now.Add(expiresIn),
		IssuedAt:          now,
		IdentityID:        identityID,
		FlowID:            uuid.NullUUID{UUID: f.ID, Valid: true},
		RecoveryAddressID: &recoveryAddressID,
	}
}

func NewRecoveryToken(identityID uuid.UUID, expiresIn time.Duration) *RecoveryToken {
	now := time.Now().UTC()
	return &RecoveryToken{
		ID:         x.NewUUID(),
		Token:      randx.MustString(32, randx.AlphaNum),
		ExpiresAt:  now.Add(expiresIn),
		IssuedAt:   now,
		IdentityID: identityID,
	}
}

func (f *RecoveryToken) Valid() error {
	if f.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(flow.NewFlowExpiredError(f.ExpiresAt))
	}
	return nil
}
