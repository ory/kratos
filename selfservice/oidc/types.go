package oidc

import "github.com/ory/kratos/selfservice"

type CredentialsConfig struct {
	Subject  string `json:"subject"`
	Provider string `json:"provider"`
}

type RequestMethodConfigProvider struct {
	Fields selfservice.FormFields `json:"fields,omitempty"`
}

type RequestMethodConfig struct {
	Action    string                  `json:"action"`
	Errors    []selfservice.FormError `json:"errors,omitempty"`
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

func (c *RequestMethodConfig) AddError(err *selfservice.FormError) {
	c.Errors = append(c.Errors, *err)
}

func (c *RequestMethodConfig) Reset() {
	c.Errors = nil
	c.Fields.Reset()
}

func (c *RequestMethodConfig) GetFormFields() selfservice.FormFields {
	return c.Fields
}
