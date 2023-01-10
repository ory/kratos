// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ory/x/httpx"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/stringsx"

	"github.com/pkg/errors"

	"github.com/gofrs/uuid"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
	"github.com/ory/x/randx"
)

var ErrIdentityDisabled = herodot.ErrUnauthorized.WithError("identity is disabled").WithReason("This account was disabled.")

type lifespanProvider interface {
	SessionLifespan(ctx context.Context) time.Duration
}

type refreshWindowProvider interface {
	SessionRefreshMinTimeLeft(ctx context.Context) time.Duration
}

// Device corresponding to a Session
//
// swagger:model sessionDevice
type Device struct {
	// Device record ID
	//
	// required: true
	ID uuid.UUID `json:"id" faker:"-" db:"id"`

	// SessionID is a helper struct field for gobuffalo.pop.
	SessionID uuid.UUID `json:"-" faker:"-" db:"session_id"`

	// IPAddress of the client
	IPAddress *string `json:"ip_address" faker:"ptr_ipv4" db:"ip_address"`

	// UserAgent of the client
	UserAgent *string `json:"user_agent" faker:"-" db:"user_agent"`

	// Geo Location corresponding to the IP Address
	Location *string `json:"location" faker:"ptr_geo_location" db:"location"`

	// Time of capture
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`

	// Last updated at
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`

	NID uuid.UUID `json:"-"  faker:"-" db:"nid"`
}

func (m Device) TableName(ctx context.Context) string {
	return "session_devices"
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

	// Devices has history of all endpoints where the session was used
	Devices []Device `json:"devices" faker:"-" has_many:"session_devices" fk_id:"session_id"`

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

func (s Session) PageToken() keysetpagination.PageToken {
	return keysetpagination.StringPageToken(s.ID.String())
}

func (s Session) DefaultPageToken() keysetpagination.PageToken {
	return keysetpagination.StringPageToken(uuid.Nil.String())
}

func (s Session) TableName(ctx context.Context) string {
	return "sessions"
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
			case identity.CredentialsTypeRecoveryCode:
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

func NewActiveSession(r *http.Request, i *identity.Identity, c lifespanProvider, authenticatedAt time.Time, completedLoginFor identity.CredentialsType, completedLoginAAL identity.AuthenticatorAssuranceLevel) (*Session, error) {
	s := NewInactiveSession()
	s.CompletedLoginFor(completedLoginFor, completedLoginAAL)
	if err := s.Activate(r, i, c, authenticatedAt); err != nil {
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

func (s *Session) Activate(r *http.Request, i *identity.Identity, c lifespanProvider, authenticatedAt time.Time) error {
	if i != nil && !i.IsActive() {
		return ErrIdentityDisabled.WithDetail("identity_id", i.ID)
	}

	s.Active = true
	s.ExpiresAt = authenticatedAt.Add(c.SessionLifespan(r.Context()))
	s.AuthenticatedAt = authenticatedAt
	s.IssuedAt = authenticatedAt
	s.Identity = i
	s.IdentityID = i.ID

	s.SaveSessionDeviceInformation(r)
	s.SetAuthenticatorAssuranceLevel()
	return nil
}

func (s *Session) SaveSessionDeviceInformation(r *http.Request) {
	var device Device

	device.ID = x.NewUUID()
	device.SessionID = s.ID

	agent := r.Header["User-Agent"]
	if len(agent) > 0 {
		device.UserAgent = stringsx.GetPointer(strings.Join(agent, " "))
	}

	device.IPAddress = stringsx.GetPointer(httpx.ClientIP(r))

	var clientGeoLocation []string
	if r.Header.Get("Cf-Ipcity") != "" {
		clientGeoLocation = append(clientGeoLocation, r.Header.Get("Cf-Ipcity"))
	}
	if r.Header.Get("Cf-Ipcountry") != "" {
		clientGeoLocation = append(clientGeoLocation, r.Header.Get("Cf-Ipcountry"))
	}
	device.Location = stringsx.GetPointer(strings.Join(clientGeoLocation, ", "))

	s.Devices = append(s.Devices, device)
}

func (s *Session) Declassify() *Session {
	s.Identity = s.Identity.CopyWithoutCredentials()
	return s
}

func (s *Session) IsActive() bool {
	return s.Active && s.ExpiresAt.After(time.Now()) && (s.Identity == nil || s.Identity.IsActive())
}

func (s *Session) Refresh(ctx context.Context, c lifespanProvider) *Session {
	s.ExpiresAt = time.Now().Add(c.SessionLifespan(ctx)).UTC()
	return s
}

func (s *Session) CanBeRefreshed(ctx context.Context, c refreshWindowProvider) bool {
	return s.ExpiresAt.Add(-c.SessionRefreshMinTimeLeft(ctx)).Before(time.Now())
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
