package template

import (
	"path/filepath"

	"github.com/ory/kratos/driver/configuration"
)

type (
	VerifyInvalid struct {
		c configuration.Provider
		m *VerifyInvalidModel
	}
	VerifyInvalidModel struct {
		To string
	}
)

func NewVerifyInvalid(c configuration.Provider, m *VerifyInvalidModel) *VerifyInvalid {
	return &VerifyInvalid{c: c, m: m}
}

func (t *VerifyInvalid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *VerifyInvalid) EmailSubject() (string, error) {
	return loadTextTemplate(filepath.Join(t.c.CourierTemplatesRoot(), "verify/invalid/email.subject.gotmpl"), t.m)
}

func (t *VerifyInvalid) EmailBody() (string, error) {
	return loadTextTemplate(filepath.Join(t.c.CourierTemplatesRoot(), "verify/invalid/email.body.gotmpl"), t.m)
}
