package identity

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/ory/gojsonschema"

	"github.com/ory/kratos/schema"
)

var _ ValidationExtender = new(ValidationExtensionIdentifier)

type ValidationExtensionIdentifier struct {
	l sync.Mutex
	v []string
	i *Identity
}

func NewValidationExtensionIdentifier() *ValidationExtensionIdentifier {
	return &ValidationExtensionIdentifier{
		v: make([]string, 0),
	}
}

func (e *ValidationExtensionIdentifier) WithIdentity(i *Identity) ValidationExtender {
	e.i = i
	return e
}

func (e *ValidationExtensionIdentifier) Call(value interface{}, config *schema.Extension, context *gojsonschema.JsonContext) error {
	if config.Credentials.Password.Identifier {
		e.l.Lock()
		defer e.l.Unlock()

		vs, ok := value.(string)
		if !ok {
			return errors.Errorf(`schema: expected value of "%s" to be string but got: %T`, context.String("."), value)
		}

		cred, ok := e.i.GetCredentials(CredentialsTypePassword)
		if !ok {
			cred = &Credentials{
				ID:          CredentialsTypePassword,
				Identifiers: []string{},
				Config:      json.RawMessage{},
			}
		}

		e.v = append(e.v, strings.ToLower(vs))
		cred.Identifiers = e.v
		e.i.SetCredentials(CredentialsTypePassword, *cred)
	}

	return nil
}
