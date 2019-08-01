package oidc

import "github.com/ory/hive/selfservice"

type CredentialsConfig struct {
	Subject  string `json:"subject"`
	Provider string `json:"provider"`
}

type RequestMethodConfigProvider struct {
	Fields selfservice.FormFields `json:"fields,omitempty"`
}

type RequestMethodConfig struct {
	Action    string                  `json:"action"`
	Error     string                  `json:"error,omitempty"`
	Fields    selfservice.FormFields  `json:"fields,omitempty"`
	Providers []selfservice.FormField `json:"providers"`
}

func NewRequestMethodConfig() *RequestMethodConfig {
	return &RequestMethodConfig{
		Fields: selfservice.FormFields{},
	}
}

type request interface {
	Valid() error
	GetID() string
}

func (c *RequestMethodConfig) SetError(err string) {
	c.Error = err
}

func (c *RequestMethodConfig) Reset() {
	c.Error = ""
	c.Fields.Reset()
}

func (c *RequestMethodConfig) GetFormFields() selfservice.FormFields {
	return c.Fields
}
