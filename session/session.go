package session

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/herodot"
	"github.com/ory/kratos/corp"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
	"github.com/ory/x/randx"
)

var ErrIdentityDisabled = herodot.ErrUnauthorized.WithError("identity is disabled").WithReason("This account was disabled.")

// A Session
//
// swagger:model session
type Session struct {
	// Session ID
	//
	// required: true
	ID uuid.UUID `json:"id" faker:"-" db:"id"`

	// Whether or not the session is active.
	Active bool `json:"active" db:"active"`

	// The Session Expiry
	//
	// When this session expires at.
	ExpiresAt time.Time `json:"expires_at" db:"expires_at" faker:"time_type"`

	// The Session Authentication Timestamp
	//
	// When this session was authenticated at.
	AuthenticatedAt time.Time `json:"authenticated_at" db:"authenticated_at" faker:"time_type"`

	// The Session Issuance Timestamp
	//
	// When this session was authenticated at.
	IssuedAt time.Time `json:"issued_at" db:"issued_at" faker:"time_type"`

	// required: true
	Identity *identity.Identity `json:"identity" faker:"identity" db:"-" belongs_to:"identities" fk_id:"IdentityID"`

	// IdentityID is a helper struct field for gobuffalo.pop.
	IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`

	// The Session Token
	//
	// The token of this session.
	Token     string    `json:"-" db:"token"`
	NID       uuid.UUID `json:"-"  faker:"-" db:"nid"`
}

func (s Session) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "sessions")
}

func NewActiveSession(i *identity.Identity, c interface {
	SessionLifespan() time.Duration
}, authenticatedAt time.Time) (*Session, error) {
	if i != nil && !i.IsActive() {
		return nil, ErrIdentityDisabled
	}

	return &Session{
		ID:              x.NewUUID(),
		ExpiresAt:       authenticatedAt.Add(c.SessionLifespan()),
		AuthenticatedAt: authenticatedAt,
		IssuedAt:        time.Now().UTC(),
		Identity:        i,
		IdentityID:      i.ID,
		Token:           randx.MustString(32, randx.AlphaNum),
		Active:          true,
	}, nil
}

type Device struct {
	UserAgent string      `json:"user_agent"`
	SeenAt    []time.Time `json:"seen_at" faker:"time_types"`
}

func (s *Session) Declassify() *Session {
	s.Identity = s.Identity.CopyWithoutCredentials()
	return s
}

func (s *Session) IsActive() bool {
	return s.Active && s.ExpiresAt.After(time.Now()) && (s.Identity == nil || s.Identity.IsActive())
}
