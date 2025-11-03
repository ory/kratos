// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/wI2L/jsondiff"

	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/sqlxx"
)

// Authenticator Assurance Level (AAL)
//
// The authenticator assurance level can be one of "aal1", "aal2", or "aal3". A higher number means that it is harder
// for an attacker to compromise the account.
//
// Generally, "aal1" implies that one authentication factor was used while AAL2 implies that two factors (e.g.
// password + TOTP) have been used.
//
// To learn more about these levels please head over to: https://www.ory.sh/kratos/docs/concepts/credentials
//
// swagger:model authenticatorAssuranceLevel
type AuthenticatorAssuranceLevel string

const (
	NoAuthenticatorAssuranceLevel AuthenticatorAssuranceLevel = "aal0"
	AuthenticatorAssuranceLevel1  AuthenticatorAssuranceLevel = "aal1"
	AuthenticatorAssuranceLevel2  AuthenticatorAssuranceLevel = "aal2"
)

type NullableAuthenticatorAssuranceLevel struct {
	sql.NullString
}

// NewNullableAuthenticatorAssuranceLevel returns a new NullableAuthenticatorAssuranceLevel
func NewNullableAuthenticatorAssuranceLevel(aal AuthenticatorAssuranceLevel) NullableAuthenticatorAssuranceLevel {
	switch aal {
	case NoAuthenticatorAssuranceLevel:
		fallthrough
	case AuthenticatorAssuranceLevel1:
		fallthrough
	case AuthenticatorAssuranceLevel2:
		return NullableAuthenticatorAssuranceLevel{sql.NullString{
			String: string(aal),
			Valid:  true,
		}}
	default:
		return NullableAuthenticatorAssuranceLevel{sql.NullString{}}
	}
}

// ToAAL returns the AuthenticatorAssuranceLevel value of the given NullableAuthenticatorAssuranceLevel.
func (n NullableAuthenticatorAssuranceLevel) ToAAL() (AuthenticatorAssuranceLevel, bool) {
	if !n.Valid {
		return "", false
	}
	switch n.String {
	case string(NoAuthenticatorAssuranceLevel):
		return NoAuthenticatorAssuranceLevel, true
	case string(AuthenticatorAssuranceLevel1):
		return AuthenticatorAssuranceLevel1, true
	case string(AuthenticatorAssuranceLevel2):
		return AuthenticatorAssuranceLevel2, true
	default:
		return "", false
	}
}

// CredentialsType  represents several different credential types, like password credentials, passwordless credentials,
// and so on.
//
// swagger:enum CredentialsType
type CredentialsType string

// Please make sure to add all of these values to the test that ensures they are created during migration
const (
	CredentialsTypePassword CredentialsType = "password"
	CredentialsTypeOIDC     CredentialsType = "oidc"
	CredentialsTypeTOTP     CredentialsType = "totp"
	CredentialsTypeLookup   CredentialsType = "lookup_secret"
	CredentialsTypeWebAuthn CredentialsType = "webauthn"
	CredentialsTypeCodeAuth CredentialsType = "code"
	CredentialsTypePasskey  CredentialsType = "passkey"
	CredentialsTypeProfile  CredentialsType = "profile"
	CredentialsTypeSAML     CredentialsType = "saml"
)

func (c CredentialsType) String() string {
	return string(c)
}

func (c CredentialsType) ToUiNodeGroup() node.UiNodeGroup {
	switch c {
	case CredentialsTypePassword:
		return node.PasswordGroup
	case CredentialsTypeOIDC, CredentialsTypeSAML:
		return node.OpenIDConnectGroup
	case CredentialsTypeTOTP:
		return node.TOTPGroup
	case CredentialsTypeWebAuthn:
		return node.WebAuthnGroup
	case CredentialsTypeLookup:
		return node.LookupGroup
	case CredentialsTypeCodeAuth:
		return node.CodeGroup
	case CredentialsTypePasskey:
		return node.PasskeyGroup
	default:
		return node.DefaultGroup
	}
}

var AllCredentialTypes = []CredentialsType{
	CredentialsTypePassword,
	CredentialsTypeOIDC,
	// CredentialsTypeSAML, placeholder for the OEL version
	CredentialsTypeTOTP,
	CredentialsTypeLookup,
	CredentialsTypeWebAuthn,
	CredentialsTypeCodeAuth,
	CredentialsTypePasskey,
}

const (
	// CredentialsTypeRecoveryLink is a special credential type linked to the link strategy (recovery flow).
	// It is not used within the credentials object itself.
	CredentialsTypeRecoveryLink CredentialsType = "link_recovery"
	CredentialsTypeRecoveryCode CredentialsType = "code_recovery"
)

// ParseCredentialsType parses a string into a CredentialsType or returns false as the second argument.
func ParseCredentialsType(in string) (CredentialsType, bool) {
	switch t := CredentialsType(in); t {
	case CredentialsTypePassword,
		CredentialsTypeOIDC,
		CredentialsTypeSAML,
		CredentialsTypeTOTP,
		CredentialsTypeLookup,
		CredentialsTypeWebAuthn,
		CredentialsTypeCodeAuth,
		CredentialsTypeRecoveryLink,
		CredentialsTypeRecoveryCode,
		CredentialsTypePasskey:
		return t, true
	}
	return "", false
}

// Credentials represents a specific credential type
//
// swagger:model identityCredentials
type Credentials struct {
	ID uuid.UUID `json:"-" db:"id"`

	// Type discriminates between different types of credentials.
	Type                     CredentialsType `json:"type" db:"-"`
	IdentityCredentialTypeID uuid.UUID       `json:"-" db:"identity_credential_type_id"`

	// Identifiers represents a list of unique identifiers this credential type matches.
	Identifiers []string `json:"identifiers" db:"-"`

	// Config contains the concrete credential payload. This might contain the bcrypt-hashed password, the email
	// for passwordless authentication or access_token and refresh tokens from OpenID Connect flows.
	Config sqlxx.JSONRawMessage `json:"config,omitempty" db:"config"`

	// Version refers to the version of the credential. Useful when changing the config schema.
	Version int `json:"version" db:"version"`

	IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	NID       uuid.UUID `json:"-"  faker:"-" db:"nid"`
}

func (c Credentials) TableName(context.Context) string {
	return "identity_credentials"
}

func (c Credentials) GetID() uuid.UUID {
	return c.ID
}

// Signature returns a unique string signature for the credential.
func (c Credentials) Signature() string {
	sortedIdentifiers := slices.Clone(c.Identifiers)
	slices.Sort(sortedIdentifiers)
	identifiersStr := strings.Join(sortedIdentifiers, ",")

	// Normalize JSON config to remove whitespace and key ordering differences
	var normalizedConfig any
	if len(c.Config) > 0 {
		if err := json.Unmarshal(c.Config, &normalizedConfig); err != nil {
			// there is not much we can do when unmarshal fails except use the raw value
			normalizedConfig = c.Config
		}
	}
	return fmt.Sprintf("%v|%v|%d|%+v|%v|%v", c.Type, identifiersStr, c.Version, normalizedConfig, c.IdentityID, c.NID)
}

type (
	// swagger:ignore
	CredentialIdentifier struct {
		ID         uuid.UUID `db:"id"`
		Identifier string    `db:"identifier"`
		// IdentityCredentialsID is a helper struct field for gobuffalo.pop.
		IdentityCredentialsID uuid.UUID `json:"-" db:"identity_credential_id"`
		// IdentityCredentialsTypeID is a helper struct field for gobuffalo.pop.
		IdentityCredentialsTypeID uuid.UUID `json:"-" db:"identity_credential_type_id"`
		// CreatedAt is a helper struct field for gobuffalo.pop.
		CreatedAt time.Time `json:"created_at" db:"created_at"`
		// UpdatedAt is a helper struct field for gobuffalo.pop.
		UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
		NID       uuid.UUID `json:"-"  faker:"-" db:"nid"`
	}

	// swagger:ignore
	CredentialsTypeTable struct {
		ID   uuid.UUID       `json:"-" db:"id"`
		Name CredentialsType `json:"-" db:"name"`
	}

	// swagger:ignore
	ActiveCredentialsCounter interface {
		ID() CredentialsType
		CountActiveFirstFactorCredentials(context.Context, map[CredentialsType]Credentials) (int, error)
		CountActiveMultiFactorCredentials(context.Context, map[CredentialsType]Credentials) (int, error)
	}

	// swagger:ignore
	ActiveCredentialsCounterStrategyProvider interface {
		ActiveCredentialsCounterStrategies(context.Context) []ActiveCredentialsCounter
	}
)

func (c CredentialsTypeTable) TableName(context.Context) string {
	return "identity_credential_types"
}

func (c CredentialIdentifier) TableName(context.Context) string {
	return "identity_credential_identifiers"
}

func CredentialsEqual(a, b map[CredentialsType]Credentials) bool {
	if len(a) != len(b) {
		return false
	}

	if len(a) == 0 && len(b) == 0 {
		return true
	}

	for k, expect := range b {
		actual, found := a[k]
		if !found {
			return false
		}

		// Try to normalize configs (remove spaces etc).
		patch, err := jsondiff.CompareJSON(actual.Config, expect.Config)
		if err != nil {
			return false
		}

		if len(patch) > 0 {
			return false
		}

		expectIdentifiers, actualIdentifiers := make(map[string]struct{}, len(expect.Identifiers)), make(map[string]struct{}, len(actual.Identifiers))
		for _, i := range expect.Identifiers {
			expectIdentifiers[i] = struct{}{}
		}
		for _, i := range actual.Identifiers {
			actualIdentifiers[i] = struct{}{}
		}
		if !reflect.DeepEqual(expectIdentifiers, actualIdentifiers) {
			return false
		}
	}

	return true
}
