package identity

import (
	"context"
	"reflect"
	"time"

	"github.com/ory/kratos/corp"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlxx"
)

// CredentialsType  represents several different credential types, like password credentials, passwordless credentials,
// and so on.
type CredentialsType string

func (c CredentialsType) String() string {
	return string(c)
}

const (
	// make sure to add all of these values to the test that ensures they are created during migration
	CredentialsTypePassword CredentialsType = "password"
	CredentialsTypeOIDC     CredentialsType = "oidc"
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

	// Config contains the concrete credential payload. This might contain the bcrypt-hashed password, or the email
	// for passwordless authentication.
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
