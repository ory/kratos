package identity

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// Identity represents a kratos identity
//
// An identity can be a real human, a service, an IoT device - everything that
// can be described as an "actor" in a system.
//
// swagger:model identity
type Identity struct {
	l *sync.RWMutex

	// ID is a unique identifier chosen by you. It can be a URN (e.g. "arn:aws:iam::123456789012"),
	// a stringified integer (e.g. "123456789012"), a uuid (e.g. "9f425a8d-7efc-4768-8f23-7647a74fdf13"). It is up to you
	// to pick a format you'd like. It is discouraged to use a personally identifiable value here, like the username
	// or the email, as this field is immutable.
	ID string `json:"id" faker:"uuid_hyphenated" form:"id" db:"id"`

	Credentials map[CredentialsType]Credentials `json:"credentials,omitempty" faker:"-" db:"-"`

	TraitsSchemaURL string          `json:"traits_schema_url,omitempty" form:"-" faker:"-" db:"traits_schema_url"`
	Traits          json.RawMessage `json:"traits" form:"traits" faker:"-" db:"traits"`
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

func (i *Identity) CopyWithoutCredentials() *Identity {
	var ii = *i
	ii.Credentials = nil
	return &ii
}

func NewIdentity(traitsSchemaURL string) *Identity {
	return &Identity{
		ID:              uuid.New().String(),
		Credentials:     map[CredentialsType]Credentials{},
		Traits:          json.RawMessage("{}"),
		TraitsSchemaURL: traitsSchemaURL,
		l:               new(sync.RWMutex),
	}
}
