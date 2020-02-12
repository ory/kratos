package identity

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/kratos/schema"
)

type SchemaExtensionCredentials struct {
	i *Identity
	v []string
	l sync.Mutex
}

func NewSchemaExtensionCredentials(i *Identity) *SchemaExtensionCredentials {
	return &SchemaExtensionCredentials{i: i}
}

func (r *SchemaExtensionCredentials) Runner(_ jsonschema.ValidationContext, s schema.ExtensionConfig, value interface{}) error {
	r.l.Lock()
	defer r.l.Unlock()
	if s.Credentials.Password.Identifier {
		cred, ok := r.i.GetCredentials(CredentialsTypePassword)
		if !ok {
			cred = &Credentials{
				Type:        CredentialsTypePassword,
				Identifiers: []string{},
				Config:      json.RawMessage{},
			}
		}

		r.v = append(r.v, strings.ToLower(fmt.Sprintf("%s", value)))
		cred.Identifiers = r.v
		r.i.SetCredentials(CredentialsTypePassword, *cred)
	}
	return nil
}
