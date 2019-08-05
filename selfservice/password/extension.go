package password

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/ory/gojsonschema"

	"github.com/ory/hive/identity"
	"github.com/ory/hive/schema"
)

var _ identity.ValidationExtender = new(ValidationExtension)

type ValidationExtension struct {
	l sync.Mutex
	v []string
	i *identity.Identity
}

func NewValidationExtension() *ValidationExtension {
	return &ValidationExtension{v: make([]string, 0)}
}

func (e *ValidationExtension) WithIdentity(i *identity.Identity) identity.ValidationExtender {
	e.i = i
	return e
}

func (e *ValidationExtension) Call(value interface{}, config *schema.Extension, context *gojsonschema.JsonContext) error {
	if config.Credentials.Password.Identifier {
		e.l.Lock()
		defer e.l.Unlock()

		vs, ok := value.(string)
		if !ok {
			return errors.Errorf(`schema: expected value of "%s" to be string but got: %T`, context.String("."), value)
		}

		cred, ok := e.i.GetCredentials(CredentialsType)
		if !ok {
			cred = &identity.Credentials{
				ID:          CredentialsType,
				Identifiers: []string{},
				Options:     json.RawMessage{},
			}
		}

		cred.Identifiers = append(cred.Identifiers, strings.ToLower(vs))
		e.i.SetCredentials(CredentialsType, *cred)
	}
	return nil
}
