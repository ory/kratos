package identity

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/gofrs/uuid"
)

// CredentialsType  represents several different credential types, like password credentials, passwordless credentials,
// and so on.
type CredentialsType string

func (c CredentialsType) String() string {
	return string(c)
}

const (
	CredentialsTypePassword CredentialsType = "password"
	CredentialsTypeOIDC     CredentialsType = "oidc"
)

type (
	// Credentials represents a specific credential type
	//
	// swagger:model identityCredentials
	Credentials struct {
		ID uuid.UUID `json:"-" db:"id"`

		CredentialTypeID uuid.UUID `json:"-" db:"identity_credential_type_id"`

		// Type discriminates between different types of credentials.
		Type CredentialsType `json:"type" db:"-"`

		// Identifiers represents a list of unique identifiers this credential type matches.
		Identifiers []string `json:"identifiers" db:"-"`

		// Config contains the concrete credential payload. This might contain the bcrypt-hashed password, or the email
		// for passwordless authentication.
		Config json.RawMessage `json:"config" db:"config"`

		IdentityID                     uuid.UUID                      `json:"-" faker:"-" db:"identity_id"`
		CredentialIdentifierCollection CredentialIdentifierCollection `json:"-" faker:"-" has_many:"identity_credential_identifiers" fk_id:"identity_credential_id"`
		// CreatedAt is a helper struct field for gobuffalo.pop.
		CreatedAt time.Time `json:"-" db:"created_at"`
		// UpdatedAt is a helper struct field for gobuffalo.pop.
		UpdatedAt time.Time `json:"-" db:"updated_at"`
	}

	// swagger:ignore
	CredentialIdentifier struct {
		ID         uuid.UUID `db:"id"`
		Identifier string    `db:"identifier"`
		// IdentityCredentialsID is a helper struct field for gobuffalo.pop.
		IdentityCredentialsID uuid.UUID `json:"-" db:"identity_credential_id"`
		// CreatedAt is a helper struct field for gobuffalo.pop.
		CreatedAt time.Time `json:"-" db:"created_at"`
		// UpdatedAt is a helper struct field for gobuffalo.pop.
		UpdatedAt time.Time `json:"-" db:"updated_at"`
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
)

func (c CredentialsTypeTable) TableName() string {
	return "identity_credential_types"
}

func (c CredentialsCollection) TableName() string {
	return "identity_credentials"
}

func (c Credentials) TableName() string {
	return "identity_credentials"
}

func (c CredentialIdentifierCollection) TableName() string {
	return "identity_credential_identifiers"
}

func (c CredentialIdentifier) TableName() string {
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

		if string(expect.Config) != string(actual.Config) {
			return false
		}

		if !reflect.DeepEqual(expect.Identifiers, actual.Identifiers) {
			return false
		}
	}

	return true
}
