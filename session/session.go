package session

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

type Session struct {
	ID              uuid.UUID          `json:"sid" faker:"uuid" db:"id"`
	ExpiresAt       time.Time          `json:"expires_at" db:"expires_at" faker:"time_type"`
	AuthenticatedAt time.Time          `json:"authenticated_at" db:"authenticated_at" faker:"time_type"`
	IssuedAt        time.Time          `json:"issued_at" db:"issued_at" faker:"time_type"`
	Identity        *identity.Identity `json:"identity" faker:"identity" db:"-" belongs_to:"identities" fk_id:"IdentityID"`

	// IdentityID is a helper struct field for gobuffalo.pop.
	IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`
	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`

	modifiedIdentity bool `faker:"-" db:"-"`
}

func (s Session) TableName() string {
	return "sessions"
}

func NewSession(i *identity.Identity, r *http.Request, c interface {
	SessionLifespan() time.Duration
}) *Session {
	return &Session{
		ID:        x.NewUUID(),
		ExpiresAt: time.Now().UTC().Add(c.SessionLifespan()),
		IssuedAt:  time.Now().UTC(),
		Identity:  i,
	}
}

type Device struct {
	UserAgent string      `json:"user_agent"`
	SeenAt    []time.Time `json:"seen_at" faker:"time_types"`
}

func (s *Session) UpdateIdentity(i *identity.Identity) *Session {
	s.Identity = i
	s.modifiedIdentity = true
	return s
}

func (s *Session) GetIdentity() *identity.Identity {
	return s.Identity
}

func (s *Session) WasIdentityModified() bool {
	return s.modifiedIdentity
}

func (s *Session) ResetModifiedIdentityFlag() *Session {
	s.modifiedIdentity = false
	return s
}
