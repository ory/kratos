package link

import (
	"context"
	"time"

	"github.com/ory/kratos/corp"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/randx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/x"
)

type VerificationToken struct {
	// ID represents the tokens's unique ID.
	//
	// required: true
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// Token represents the verification token. It can not be longer than 64 chars!
	Token string `json:"-" db:"token"`

	// VerifiableAddress links this token to a verification address.
	// required: true
	VerifiableAddress *identity.VerifiableAddress `json:"verification_address" belongs_to:"identity_verifiable_addresses" fk_id:"VerificationAddVerifiableAddressIDressID"`

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
	// VerifiableAddressID is a helper struct field for gobuffalo.pop.
	VerifiableAddressID uuid.UUID `json:"-" faker:"-" db:"identity_verifiable_address_id"`
	// FlowID is a helper struct field for gobuffalo.pop.
	FlowID uuid.NullUUID `json:"-" faker:"-" db:"selfservice_verification_flow_id"`
	NID    uuid.UUID     `json:"-"  faker:"-" db:"nid"`
}

func (VerificationToken) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "identity_verification_tokens")
}

func NewSelfServiceVerificationToken(address *identity.VerifiableAddress, f *verification.Flow, expiresIn time.Duration) *VerificationToken {
	now := time.Now().UTC()
	return &VerificationToken{
		ID:                x.NewUUID(),
		Token:             randx.MustString(32, randx.AlphaNum),
		VerifiableAddress: address,
		ExpiresAt:         now.Add(expiresIn),
		IssuedAt:          now,
		FlowID:            uuid.NullUUID{UUID: f.ID, Valid: true}}
}

func (f *VerificationToken) Valid() error {
	if f.ExpiresAt.Before(time.Now().UTC()) {
		return errors.WithStack(verification.NewFlowExpiredError(f.ExpiresAt))
	}
	return nil
}
