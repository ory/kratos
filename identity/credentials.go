package identity

import (
	"context"
	"reflect"
	"time"

	"github.com/ory/kratos/corp"

	"github.com/gofrs/uuid"

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
	AuthenticatorAssuranceLevel3  AuthenticatorAssuranceLevel = "aal3"
)

// CredentialsType  represents several different credential types, like password credentials, passwordless credentials,
// and so on.
//
// swagger:model identityCredentialsType
type CredentialsType string

func (c CredentialsType) String() string {
	return string(c)
}

// Please make sure to add all of these values to the test that ensures they are created during migration
const (
	CredentialsTypePassword CredentialsType = "password"
	CredentialsTypeOIDC     CredentialsType = "oidc"
	CredentialsTypeTOTP     CredentialsType = "totp"
	CredentialsTypeLookup   CredentialsType = "lookup_secret"
	CredentialsTypeWebAuthn CredentialsType = "webauthn"
)

const (
	// CredentialsTypeRecoveryLink is a special credential type linked to the link strategy (recovery flow).
	// It is not used within the credentials object itself.
	CredentialsTypeRecoveryLink CredentialsType = "link_recovery"
)

// Credentials represents a specific credential type
//
// swagger:model identityCredentials
type Credentials struct {
	ID uuid.UUID `json:"-" db:"id"`

	CredentialTypeID uuid.UUID `json:"-" db:"identity_credential_type_id"`

	// Type discriminates between different types of credentials.
	Type CredentialsType `json:"type" db:"-"`

	// Identifiers represents a list of unique identifiers this credential type matches.
	Identifiers []string `json:"identifiers" db:"-"`

	// Config contains the concrete credential payload. This might contain the bcrypt-hashed password, the email
	// for passwordless authentication or access_token and refresh tokens from OpenID Connect flows.
	Config sqlxx.JSONRawMessage `json:"config,omitempty" db:"config"`

	IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	NID       uuid.UUID `json:"-"  faker:"-" db:"nid"`
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
	CredentialsCollection []Credentials

	// swagger:ignore
	CredentialIdentifierCollection []CredentialIdentifier

	// swagger:ignore
	ActiveCredentialsCounter interface {
		ID() CredentialsType
		CountActiveCredentials(cc map[CredentialsType]Credentials) (int, error)
	}

	// swagger:ignore
	ActiveCredentialsCounterStrategyProvider interface {
		ActiveCredentialsCounterStrategies(context.Context) []ActiveCredentialsCounter
	}
)

func (c CredentialsTypeTable) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "identity_credential_types")
}

func (c CredentialsCollection) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "identity_credentials")
}

func (c Credentials) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "identity_credentials")
}

func (c CredentialIdentifierCollection) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "identity_credential_identifiers")
}

func (c CredentialIdentifier) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "identity_credential_identifiers")
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

		if string(expect.Config) != string(actual.Config) {
			return false
		}

		if !reflect.DeepEqual(expect.Identifiers, actual.Identifiers) {
			return false
		}
	}

	return true
}
