package session

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ory/hive-cloud/hive/identity"
)

type Session struct {
	SID             string             `json:"sid"`
	ExpiresAt       time.Time          `json:"expires_at" faker:"time_type"`
	AuthenticatedAt time.Time          `json:"authenticated_at" faker:"time_type"`
	IssuedAt        time.Time          `json:"issued_at" faker:"time_type"`
	Identity        *identity.Identity `json:"identity"`
	Devices         []Device           `json:"devices,omitempty" faker:"-"`

	modifiedIdentity bool
}

func NewSession(i *identity.Identity, r *http.Request, c Configuration) *Session {
	return &Session{
		SID:       uuid.New().String(),
		ExpiresAt: time.Now().UTC().Add(c.SessionLifespan()),
		IssuedAt:  time.Now().UTC(),
		Identity:  i,
		Devices: []Device{
			{
				// IP: r.RemoteAddr,
				UserAgent: r.UserAgent(),
				SeenAt: []time.Time{
					time.Now().UTC(),
				},
			},
		},
	}
}

type Device struct {
	UserAgent string `json:"user_agent"`
	// IP string `json:"ip"`
	SeenAt []time.Time `json:"seen_at" faker:"time_types"`
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
