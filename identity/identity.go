package identity

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/tidwall/sjson"

	"github.com/tidwall/gjson"

	"github.com/ory/kratos/cipher"

	"github.com/ory/kratos/corp"

	"github.com/ory/herodot"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver/config"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/x"
)

// An Identity's State
//
// The state can either be `active` or `inactive`.
//
// swagger:model identityState
type State string

const (
	StateActive   State = "active"
	StateInactive State = "inactive"
)

func (lt State) IsValid() error {
	switch lt {
	case StateActive, StateInactive:
		return nil
	}
	return errors.New("identity state is not valid")
}

//type IdentifierCredential struct {
//	Subject      string `json:"subject"`
//	Provider     string `json:"provider"`
//	AccessToken  string `json:"access_token"`
//	RefreshToken string `json:"refresh_token"`
//}

// Identity represents an Ory Kratos identity
//
// An identity can be a real human, a service, an IoT device - everything that
// can be described as an "actor" in a system.
//
// swagger:model identity
type Identity struct {
	l *sync.RWMutex `db:"-" faker:"-"`

	// ID is the identity's unique identifier.
	//
	// The Identity ID can not be changed and can not be chosen. This ensures future
	// compatibility and optimization for distributed stores such as CockroachDB.
	//
	// required: true
	ID uuid.UUID `json:"id" faker:"-" db:"id"`

	// Credentials represents all credentials that can be used for authenticating this identity.
	Credentials map[CredentialsType]Credentials `json:"credentials,omitempty" faker:"-" db:"-"`

	//// IdentifierCredentials contains the access and refresh token for oidc identifier
	//IdentifierCredentials []IdentifierCredential `json:"identifier_credentials,omitempty" faker:"-" db:"-"`

	// SchemaID is the ID of the JSON Schema to be used for validating the identity's traits.
	//
	// required: true
	SchemaID string `json:"schema_id" faker:"-" db:"schema_id"`

	// SchemaURL is the URL of the endpoint where the identity's traits schema can be fetched from.
	//
	// format: url
	// required: true
	SchemaURL string `json:"schema_url" faker:"-" db:"-"`

	// State is the identity's state.
	//
	// This value has currently no effect.
	State State `json:"state" faker:"-" db:"state"`

	// StateChangedAt contains the last time when the identity's state changed.
	StateChangedAt *sqlxx.NullTime `json:"state_changed_at,omitempty" faker:"-" db:"state_changed_at"`

	// Traits represent an identity's traits. The identity is able to create, modify, and delete traits
	// in a self-service manner. The input will always be validated against the JSON Schema defined
	// in `schema_url`.
	//
	// required: true
	Traits Traits `json:"traits" faker:"-" db:"traits"`

	// VerifiableAddresses contains all the addresses that can be verified by the user.
	//
	// Extensions:
	// ---
	// x-omitempty: true
	// ---
	VerifiableAddresses []VerifiableAddress `json:"verifiable_addresses,omitempty" faker:"-" has_many:"identity_verifiable_addresses" fk_id:"identity_id"`

	// RecoveryAddresses contains all the addresses that can be used to recover an identity.
	//
	// Extensions:
	// ---
	// x-omitempty: true
	// ---
	RecoveryAddresses []RecoveryAddress `json:"recovery_addresses,omitempty" faker:"-" has_many:"identity_recovery_addresses" fk_id:"identity_id"`

	// Store metadata about the identity which the identity itself can see when calling for example the
	// session endpoint. Do not store sensitive information (e.g. credit score) about the identity in this field.
	MetadataPublic sqlxx.NullJSONRawMessage `json:"metadata_public" faker:"-" db:"metadata_public"`

	// Store metadata about the user which is only accessible through admin APIs such as `GET /admin/identities/<id>`.
	MetadataAdmin sqlxx.NullJSONRawMessage `json:"metadata_admin,omitempty" faker:"-" db:"metadata_admin"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	NID       uuid.UUID `json:"-"  faker:"-" db:"nid"`
}

// Traits represent an identity's traits. The identity is able to create, modify, and delete traits
// in a self-service manner. The input will always be validated against the JSON Schema defined
// in `schema_url`.
//
// swagger:model identityTraits
type Traits json.RawMessage

func (t *Traits) Scan(value interface{}) error {
	return sqlxx.JSONScan(t, value)
}

func (t Traits) Value() (driver.Value, error) {
	return sqlxx.JSONValue(t)
}

func (t *Traits) String() string {
	return string(*t)
}

// MarshalJSON returns m as the JSON encoding of m.
func (t Traits) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("null"), nil
	}
	return t, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (t *Traits) UnmarshalJSON(data []byte) error {
	if t == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*t = append((*t)[0:0], data...)
	return nil
}

func (i Identity) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "identities")
}

func (i *Identity) lock() *sync.RWMutex {
	if i.l == nil {
		i.l = new(sync.RWMutex)
	}
	return i.l
}

func (i *Identity) IsActive() bool {
	return i.State == StateActive
}

func (i *Identity) SetCredentials(t CredentialsType, c Credentials) {
	i.lock().Lock()
	defer i.lock().Unlock()
	if i.Credentials == nil {
		i.Credentials = make(map[CredentialsType]Credentials)
	}

	c.Type = t
	i.Credentials[t] = c
}

func (i *Identity) SetCredentialsWithConfig(t CredentialsType, c Credentials, conf interface{}) (err error) {
	i.lock().Lock()
	defer i.lock().Unlock()
	if i.Credentials == nil {
		i.Credentials = make(map[CredentialsType]Credentials)
	}

	c.Config, err = json.Marshal(conf)
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode %s credentials to JSON: %s", t, err))
	}

	c.Type = t
	i.Credentials[t] = c
	return nil
}

func (i *Identity) DeleteCredentialsType(t CredentialsType) {
	i.lock().Lock()
	defer i.lock().Unlock()
	if i.Credentials == nil {
		return
	}

	delete(i.Credentials, t)
}

func (i *Identity) GetCredentialsOr(t CredentialsType, or *Credentials) *Credentials {
	c, ok := i.GetCredentials(t)
	if !ok {
		return or
	}
	return c
}

func (i *Identity) UpsertCredentialsConfig(t CredentialsType, conf []byte, version int) {
	c, ok := i.GetCredentials(t)
	if !ok {
		c = &Credentials{}
	}

	c.Type = t
	c.IdentityID = i.ID
	c.Config = conf
	c.Version = version

	i.SetCredentials(t, *c)
}

func (i *Identity) GetCredentials(t CredentialsType) (*Credentials, bool) {
	i.lock().RLock()
	defer i.lock().RUnlock()

	if c, ok := i.Credentials[t]; ok {
		return &c, true
	}

	return nil, false
}

func (i *Identity) ParseCredentials(t CredentialsType, config interface{}) (*Credentials, error) {
	i.lock().RLock()
	defer i.lock().RUnlock()

	if c, ok := i.Credentials[t]; ok {
		if err := json.Unmarshal(c.Config, config); err != nil {
			return nil, errors.WithStack(err)
		}
		return &c, nil
	}

	return nil, herodot.ErrNotFound.WithReasonf("identity does not have credential type %s", t)
}

func (i *Identity) CopyWithoutCredentials() *Identity {
	var ii = *i
	ii.Credentials = nil
	return &ii
}

func NewIdentity(traitsSchemaID string) *Identity {
	if traitsSchemaID == "" {
		traitsSchemaID = config.DefaultIdentityTraitsSchemaID
	}

	stateChangedAt := sqlxx.NullTime(time.Now().UTC())
	return &Identity{
		ID:                  x.NewUUID(),
		Credentials:         map[CredentialsType]Credentials{},
		Traits:              Traits("{}"),
		SchemaID:            traitsSchemaID,
		VerifiableAddresses: []VerifiableAddress{},
		State:               StateActive,
		StateChangedAt:      &stateChangedAt,
		l:                   new(sync.RWMutex),
	}
}

func (i Identity) GetID() uuid.UUID {
	return i.ID
}

func (i Identity) GetNID() uuid.UUID {
	return i.NID
}

func (i Identity) MarshalJSON() ([]byte, error) {
	type localIdentity Identity
	i.Credentials = nil
	i.MetadataAdmin = nil
	result, err := json.Marshal(localIdentity(i))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (i *Identity) UnmarshalJSON(b []byte) error {
	type localIdentity Identity
	err := json.Unmarshal(b, (*localIdentity)(i))
	i.Credentials = nil
	i.MetadataAdmin = nil
	return err
}

type WithCredentialsAndAdminMetadataInJSON Identity

func (i WithCredentialsAndAdminMetadataInJSON) MarshalJSON() ([]byte, error) {
	type localIdentity Identity
	return json.Marshal(localIdentity(i))
}

type WithCredentialsMetadataAndAdminMetadataInJSON Identity

func (i WithCredentialsMetadataAndAdminMetadataInJSON) MarshalJSON() ([]byte, error) {
	type localIdentity Identity
	for k, v := range i.Credentials {
		v.Config = nil
		i.Credentials[k] = v
	}
	return json.Marshal(localIdentity(i))
}

func (i *Identity) ValidateNID() error {
	expected := i.NID
	if expected == uuid.Nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReason("Received empty nid."))
	}

	for _, r := range i.RecoveryAddresses {
		if r.NID != expected {
			return errors.WithStack(herodot.ErrInternalServerError.WithReason("Mismatching nid for recovery addresses."))
		}
	}

	for _, r := range i.VerifiableAddresses {
		if r.NID != expected {
			return errors.WithStack(herodot.ErrInternalServerError.WithReason("Mismatching nid for verifiable addresses."))
		}
	}

	for _, r := range i.Credentials {
		if r.NID != expected {
			return errors.WithStack(herodot.ErrInternalServerError.WithReason("Mismatching nid for credentials."))
		}
	}

	return nil
}

func (i *Identity) WithDeclassifiedCredentialsOIDC(ctx context.Context, c cipher.Provider) (*Identity, error) {
	credsToPublish := make(map[CredentialsType]Credentials)

	for ct, original := range i.Credentials {
		if ct != CredentialsTypeOIDC {
			toPublish := original
			toPublish.Config = []byte{}
			credsToPublish[ct] = toPublish
			continue
		}

		toPublish := original
		toPublish.Config = []byte{}

		for _, token := range []string{"initial_id_token", "initial_access_token", "initial_refresh_token"} {
			var i int
			var err error
			gjson.GetBytes(original.Config, "providers").ForEach(func(_, v gjson.Result) bool {
				key := fmt.Sprintf("%d.%s", i, token)
				ciphertext := v.Get(token).String()

				var plaintext []byte
				plaintext, err = c.Cipher().Decrypt(ctx, ciphertext)
				if err != nil {
					return false
				}

				toPublish.Config, err = sjson.SetBytes(toPublish.Config, "providers."+key, string(plaintext))
				if err != nil {
					return false
				}

				toPublish.Config, err = sjson.SetBytes(toPublish.Config, fmt.Sprintf("providers.%d.subject", i), v.Get("subject").String())
				if err != nil {
					return false
				}

				toPublish.Config, err = sjson.SetBytes(toPublish.Config, fmt.Sprintf("providers.%d.provider", i), v.Get("provider").String())
				if err != nil {
					return false
				}

				i++
				return true
			})

			if err != nil {
				return nil, err
			}
		}

		credsToPublish[ct] = toPublish
	}

	ii := *i
	ii.Credentials = credsToPublish
	return &ii, nil
}
