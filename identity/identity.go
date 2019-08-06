package identity

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// CredentialsType  represents several different credential types, like password credentials, passwordless credentials,
// and so on.
type CredentialsType string

// Identity represents a hive identity
//
// An identity can be a real human, a service, an IoT device - everything that
// can be described as an "actor" in a system.
//
// swagger:model identity
type Identity struct {
	l *sync.RWMutex

	// PK: The primary key used for hive-internal processing. It is auto-assigned and immutable.
	PK uint64 `json:"-" faker:"-" form:"-"`

	// ID is a unique identifier chosen by you. It can be a URN (e.g. "arn:aws:iam::123456789012"),
	// a stringified integer (e.g. "123456789012"), a uuid (e.g. "9f425a8d-7efc-4768-8f23-7647a74fdf13"). It is up to you
	// to pick a format you'd like. It is discouraged to use a personally identifiable value here, like the username
	// or the email, as this field is immutable.
	ID string `json:"id" faker:"uuid_hyphenated" form:"id"`

	Credentials map[CredentialsType]Credentials `json:"-" faker:"-"`

	// MetadataSchemaURL string          `json:"metadata_schema_url,omitempty" form:"-"`
	// Metadata          json.RawMessage `json:"metadata,omitempty" form:"-" faker:"-"`

	TraitsSchemaURL string          `json:"traits_schema_url,omitempty" form:"-"`
	Traits          json.RawMessage `json:"traits" form:"traits" faker:"-"`
}

func (i *Identity) lock() *sync.RWMutex {
	if i.l == nil {
		i.l = new(sync.RWMutex)
	}
	return i.l
}

func (i *Identity) SetCredentials(t CredentialsType, c Credentials) {
	i.lock().Lock()
	defer i.lock().Unlock()
	if i.Credentials == nil {
		i.Credentials = make(map[CredentialsType]Credentials)
	}

	c.ID = t
	i.Credentials[t] = c
}

func (i *Identity) GetCredentials(t CredentialsType) (*Credentials, bool) {
	i.lock().RLock()
	defer i.lock().RUnlock()

	if c, ok := i.Credentials[t]; ok {
		return &c, true
	}

	return nil, false
}

func (i *Identity) WithoutCredentials() *Identity {
	i.lock().Lock()
	defer i.lock().Unlock()
	i.Credentials = nil
	return i
}

// Credentials represents a specific credential type
//
// swagger:model identityCredentials
type Credentials struct {
	// PK: The primary key used for hive-internal processing. It is auto-assigned and immutable.
	PK uint64 `json:"-" faker:"-"`

	// RequestID discriminates between different credential types.
	ID CredentialsType `json:"id"`

	// Identifiers represents a list of unique identifiers this credential type matches.
	Identifiers []string `json:"identifiers"`

	// Options contains the concrete credential payload. This might contain the bcrypt-hashed password, or the email
	// for passwordless authentication.
	Options json.RawMessage `json:"options"`
}

func NewIdentity(traitsSchemaURL string) *Identity {
	return &Identity{
		ID:          uuid.New().String(),
		Credentials: map[CredentialsType]Credentials{},
		// Metadata:        json.RawMessage("{}"),
		Traits:          json.RawMessage("{}"),
		TraitsSchemaURL: traitsSchemaURL,
		l:               new(sync.RWMutex),
	}
}
