package template

import (
	"path/filepath"

	"github.com/ory/kratos/driver/configuration"
)

type (
	VerificationInvalid struct {
		c configuration.Provider
		m *VerificationInvalidModel
	}
	VerificationInvalidModel struct {
		To string
	}
)

func NewVerificationInvalid(c configuration.Provider, m *VerificationInvalidModel) *VerificationInvalid {
	return &VerificationInvalid{c: c, m: m}
}

func (t *VerificationInvalid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *VerificationInvalid) EmailSubject() (string, error) {
	return loadTextTemplate(filepath.Join(t.c.CourierTemplatesRoot(), "verification/invalid/email.subject.gotmpl"), t.m)
}

func (t *VerificationInvalid) EmailBody() (string, error) {
	return loadTextTemplate(filepath.Join(t.c.CourierTemplatesRoot(), "verification/invalid/email.body.gotmpl"), t.m)
}
