package oidc

import "github.com/ory/hive/selfservice"

type CredentialsConfig struct {
	Subject  string `json:"subject"`
	Provider string `json:"provider"`
}

type RequestMethodConfigProvider struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type RequestMethodConfig struct {
	Error     string                        `json:"error,omitempty"`
	Providers []RequestMethodConfigProvider `json:"providers"`
	Used      *int                          `json:"used,omitempty"`
	Fields    selfservice.FormFields        `json:"fields,omitempty"`
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
