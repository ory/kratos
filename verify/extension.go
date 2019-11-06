package verify

import (
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
)

type Extension struct {
	// i *identity.Identity
	// l sync.Mutex
}

func NewExtension() *Extension {
	return &Extension{}
}

func (e *Extension) Callback(i *identity.Identity, value interface{}, config *schema.Extension) error {
	return nil
}
