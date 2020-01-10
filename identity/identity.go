package identity

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"sync"
	"time"

	"github.com/ory/x/urlx"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/persistence/aliases"
	"github.com/ory/kratos/x"
)

type (
	// Identity represents an ORY Kratos identity
	//
	// An identity can be a real human, a service, an IoT device - everything that
	// can be described as an "actor" in a system.
	//
	// swagger:model identity
	Identity struct {
		l *sync.RWMutex `db:"-" faker:"-"`

		// ID is a unique identifier chosen by you. It can be a URN (e.g. "arn:aws:iam::123456789012"),
		// a stringified integer (e.g. "123456789012"), a uuid (e.g. "9f425a8d-7efc-4768-8f23-7647a74fdf13"). It is up to you
		// to pick a format you'd like. It is discouraged to use a personally identifiable value here, like the username
		// or the email, as this field is immutable.
		//
		// required: true
		ID uuid.UUID `json:"id" faker:"uuid" db:"id" rw:"r"`

		// Credentials represents all credentials that can be used for authenticating this identity.
		Credentials map[CredentialsType]Credentials `json:"-" faker:"-" db:"-"`

		// TraitsSchemaURL is the URL of the JSON Schema to be used for validating the identity's traits.
		//
		// format: uri
		TraitsSchemaURL string `json:"traits_schema_url,omitempty" faker:"-" db:"-"`

		// TraitsSchemaID is the ID of the JSON schema used for validating the identity's traits.
		//
		// format: uuid
		TraitsSchemaID uuid.UUID `json:"traits_schema_id" faker:"-" db:"traits_schema_id"`

		// Traits represent an identity's traits. The identity is able to create, modify, and delete traits
		// in a self-service manner. The input will always be validated against the JSON Schema defined
		// in `traits_schema_url`.
		//
		// required: true
		Traits Traits `json:"traits" form:"traits" faker:"-" db:"traits"`

		// CredentialsCollection is a helper struct field for gobuffalo.pop.
		CredentialsCollection CredentialsCollection `json:"-" faker:"-" has_many:"identity_credentials" fk_id:"identity_id"`
		// CreatedAt is a helper struct field for gobuffalo.pop.
		CreatedAt time.Time `json:"-" db:"created_at"`
		// UpdatedAt is a helper struct field for gobuffalo.pop.
		UpdatedAt time.Time `json:"-" db:"updated_at"`
	}
	Traits json.RawMessage
)

func (t *Traits) Scan(value interface{}) error {
	return aliases.JSONScan(t, value)
}

func (t *Traits) Value() (driver.Value, error) {
	return aliases.JSONValue(t)
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

func (i Identity) TableName() string {
	return "identities"
}

func (i *Identity) InjectTraitsUrl(host string) {
	i.TraitsSchemaURL = urlx.MustJoin(host, "schemas", i.TraitsSchemaID.String())
}

func (i *Identity) lock() *sync.RWMutex {
	if i.l == nil {
		i.l = new(sync.RWMutex)
	}
	return i.l
}

func (i *Identity) CredentialsEqual(c map[CredentialsType]Credentials) bool {
	if len(c) != len(i.Credentials) {
		return false
	}

	if len(c) == 0 && len(i.Credentials) == 0 {
		return true
	}

	for k, expect := range i.Credentials {
		actual, found := c[k]
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

func (i *Identity) SetCredentials(t CredentialsType, c Credentials) {
	i.lock().Lock()
	defer i.lock().Unlock()
	if i.Credentials == nil {
		i.Credentials = make(map[CredentialsType]Credentials)
	}

	c.Type = t
	i.Credentials[t] = c
}

func (i *Identity) CopyCredentials() map[CredentialsType]Credentials {
	result := make(map[CredentialsType]Credentials)
	for id, credential := range i.Credentials {
		result[id] = Credentials{
			Type:        credential.Type,
			Identifiers: append([]string{}, credential.Identifiers...),
			Config:      append([]byte{}, credential.Config...),
		}
	}
	return result
}

func (i *Identity) GetCredentials(t CredentialsType) (*Credentials, bool) {
	i.lock().RLock()
	defer i.lock().RUnlock()

	if c, ok := i.Credentials[t]; ok {
		return &c, true
	}

	return nil, false
}

func (i *Identity) CopyWithoutCredentials() *Identity {
	var ii = *i
	ii.Credentials = nil
	return &ii
}

func NewIdentity(traitsSchemaID uuid.UUID) *Identity {
	return &Identity{
		ID:             x.NewUUID(),
		Credentials:    map[CredentialsType]Credentials{},
		Traits:         Traits(json.RawMessage("{}")),
		TraitsSchemaID: traitsSchemaID,
		l:              new(sync.RWMutex),
	}
}
