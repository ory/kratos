package session

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/gofrs/uuid"

	"github.com/ory/herodot"
	"github.com/ory/kratos/corp"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
	"github.com/ory/x/randx"
)

var ErrIdentityDisabled = herodot.ErrUnauthorized.WithError("identity is disabled").WithReason("This account was disabled.")

type lifespanProvider interface {
	SessionLifespan() time.Duration
}

type refreshWindowProvider interface {
	SessionRefreshMinTimeLeft() time.Duration
}

// A Session
//
// swagger:model session
type Session struct {
	// Session ID
	//
	// required: true
	ID uuid.UUID `json:"id" faker:"-" db:"id"`

	// Active state. If false the session is no longer active.
	Active bool `json:"active" db:"active"`

	// The Session Expiry
	//
	// When this session expires at.
	ExpiresAt time.Time `json:"expires_at" db:"expires_at" faker:"time_type"`

	// The Session Authentication Timestamp
	//
	// When this session was authenticated at. If multi-factor authentication was used this
	// is the time when the last factor was authenticated (e.g. the TOTP code challenge was completed).
	AuthenticatedAt time.Time `json:"authenticated_at" db:"authenticated_at" faker:"time_type"`

	// AuthenticationMethod Assurance Level (AAL)
	//
	// The authenticator assurance level can be one of "aal1", "aal2", or "aal3". A higher number means that it is harder
	// for an attacker to compromise the account.
	//
	// Generally, "aal1" implies that one authentication factor was used while AAL2 implies that two factors (e.g.
	// password + TOTP) have been used.
	//
	// To learn more about these levels please head over to: https://www.ory.sh/kratos/docs/concepts/credentials
	AuthenticatorAssuranceLevel identity.AuthenticatorAssuranceLevel `faker:"len=4" db:"aal" json:"authenticator_assurance_level"`

	// Authentication Method References (AMR)
	//
	// A list of authentication methods (e.g. password, oidc, ...) used to issue this session.
	AMR AuthenticationMethods `db:"authentication_methods" json:"authentication_methods"`

	// The Session Issuance Timestamp
	//
	// When this session was issued at. Usually equal or close to `authenticated_at`.
	IssuedAt time.Time `json:"issued_at" db:"issued_at" faker:"time_type"`

	// The Logout Token
	//
	// Use this token to log out a user.
	LogoutToken string `json:"-" db:"logout_token"`

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
	Token string    `json:"-" db:"token"`
	NID   uuid.UUID `json:"-"  faker:"-" db:"nid"`
}

func (s Session) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "sessions")
}

func (s *Session) CompletedLoginFor(method identity.CredentialsType, aal identity.AuthenticatorAssuranceLevel) {
	s.AMR = append(s.AMR, AuthenticationMethod{Method: method, AAL: aal, CompletedAt: time.Now().UTC()})
}

func (s *Session) SetAuthenticatorAssuranceLevel() {
	if len(s.AMR) == 0 {
		// No AMR is set
		s.AuthenticatorAssuranceLevel = identity.NoAuthenticatorAssuranceLevel
	}

	var isAAL1, isAAL2 bool
	for _, amr := range s.AMR {
		switch amr.AAL {
		case identity.AuthenticatorAssuranceLevel1:
			isAAL1 = true
		case identity.AuthenticatorAssuranceLevel2:
			isAAL2 = true
		case "":
			// Sessions before Ory Kratos 0.9 did not have the AAL
			// be part of the AMR.
			switch amr.Method {
			case identity.CredentialsTypeRecoveryLink:
				isAAL1 = true
			case identity.CredentialsTypeOIDC:
				isAAL1 = true
			case "v0.6_legacy_session":
				isAAL1 = true
			case identity.CredentialsTypePassword:
				isAAL1 = true
			case identity.CredentialsTypeWebAuthn:
				isAAL2 = true
			case identity.CredentialsTypeTOTP:
				isAAL2 = true
			case identity.CredentialsTypeLookup:
				isAAL2 = true
			}
		}
	}

	if isAAL1 && isAAL2 {
		s.AuthenticatorAssuranceLevel = identity.AuthenticatorAssuranceLevel2
	} else if isAAL1 {
		s.AuthenticatorAssuranceLevel = identity.AuthenticatorAssuranceLevel1
	} else if len(s.AMR) > 0 {
		// A fallback. If an AMR is set but we did not satisfy the above, gracefully fall back to level 1.
		s.AuthenticatorAssuranceLevel = identity.AuthenticatorAssuranceLevel1
	}
}

func NewActiveSession(i *identity.Identity, c lifespanProvider, authenticatedAt time.Time, completedLoginFor identity.CredentialsType, completedLoginAAL identity.AuthenticatorAssuranceLevel) (*Session, error) {
	s := NewInactiveSession()
	s.CompletedLoginFor(completedLoginFor, completedLoginAAL)
	if err := s.Activate(i, c, authenticatedAt); err != nil {
		return nil, err
	}
	return s, nil
}

func NewInactiveSession() *Session {
	return &Session{
		ID:                          x.NewUUID(),
		Token:                       randx.MustString(32, randx.AlphaNum),
		LogoutToken:                 randx.MustString(32, randx.AlphaNum),
		Active:                      false,
		AuthenticatorAssuranceLevel: identity.NoAuthenticatorAssuranceLevel,
	}
}

func (s *Session) Activate(i *identity.Identity, c lifespanProvider, authenticatedAt time.Time) error {
	if i != nil && !i.IsActive() {
		return ErrIdentityDisabled
	}

	s.Active = true
	s.ExpiresAt = authenticatedAt.Add(c.SessionLifespan())
	s.AuthenticatedAt = authenticatedAt
	s.IssuedAt = authenticatedAt
	s.Identity = i
	s.IdentityID = i.ID

	s.SetAuthenticatorAssuranceLevel()
	return nil
}

// swagger:model sessionDevice
type Device struct {
	// UserAgent of this device
	UserAgent string `json:"user_agent"`
}

func (s *Session) Declassify() *Session {
	s.Identity = s.Identity.CopyWithoutCredentials()
	return s
}

func (s *Session) IsActive() bool {
	return s.Active && s.ExpiresAt.After(time.Now()) && (s.Identity == nil || s.Identity.IsActive())
}

func (s *Session) Refresh(c lifespanProvider) *Session {
	s.ExpiresAt = time.Now().Add(c.SessionLifespan()).UTC()
	return s
}

func (s *Session) CanBeRefreshed(c refreshWindowProvider) bool {
	return s.ExpiresAt.Add(-c.SessionRefreshMinTimeLeft()).Before(time.Now())
}

// List of (Used) AuthenticationMethods
//
// A list of authenticators which were used to authenticate the session.
//
// swagger:model sessionAuthenticationMethods
type AuthenticationMethods []AuthenticationMethod

// AuthenticationMethod identifies an authentication method
//
// A singular authenticator used during authentication / login.
//
// swagger:model sessionAuthenticationMethod
type AuthenticationMethod struct {
	// The method used in this authenticator.
	Method identity.CredentialsType `json:"method"`

	// The AAL this method introduced.
	AAL identity.AuthenticatorAssuranceLevel `json:"aal"`

	// When the authentication challenge was completed.
	CompletedAt time.Time `json:"completed_at"`
}

// Scan implements the Scanner interface.
func (n *AuthenticationMethod) Scan(value interface{}) error {
	v := fmt.Sprintf("%s", value)
	if len(v) == 0 {
		return nil
	}
	return errors.WithStack(json.Unmarshal([]byte(v), n))
}

// Value implements the driver Valuer interface.
func (n AuthenticationMethod) Value() (driver.Value, error) {
	value, err := json.Marshal(n)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return string(value), nil
}

// Scan implements the Scanner interface.
func (n *AuthenticationMethods) Scan(value interface{}) error {
	v := fmt.Sprintf("%s", value)
	if len(v) == 0 {
		return nil
	}
	return errors.WithStack(json.Unmarshal([]byte(v), n))
}

// Value implements the driver Valuer interface.
func (n AuthenticationMethods) Value() (driver.Value, error) {
	value, err := json.Marshal(n)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return string(value), nil
}
