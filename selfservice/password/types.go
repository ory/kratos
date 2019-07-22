package password

import (
	"net/http"

	"github.com/ory/hive-cloud/hive/identity"
)

const CredentialsType identity.CredentialsType = "password"

const csrfTokenName = "csrf_token"

type CredentialsConfig struct {
	HashedPassword string `json:"hashed_password"`
}

type csrfGenerator func(r *http.Request) string

type FormFields map[string]FormField

func (fs FormFields) Reset() {
	for k, f := range fs {
		f.Error = ""
		f.Value = ""
		fs[k] = f
	}
}

func (fs FormFields) SetValue(name, value string) {
	var field FormField
	if ff, ok := fs[name]; ok {
		field = ff
	}

	field.Name = name
	field.Value = value
	fs[name] = field
}

func (fs FormFields) SetError(name, err string) {
	var field FormField
	if ff, ok := fs[name]; ok {
		field = ff
	}

	field.Name = name
	field.Error = err
	fs[name] = field
}

type FormField struct {
	Name     string   `json:"name"`
	Type     string   `json:"type,omitempty"`
	Required bool     `json:"required,omitempty"`
	Value    string   `json:"value,omitempty"`
	Options  []string `json:"options,omitempty"`
	Error    string   `json:"error,omitempty"`
}
