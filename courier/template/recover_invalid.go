package template

import (
	"path/filepath"

	"github.com/ory/kratos/driver/configuration"
)

type (
	RecoverInvalid struct {
		c configuration.Provider
		m *RecoverInvalidModel
	}
	RecoverInvalidModel struct {
		To string
	}
)

func NewRecoverInvalid(c configuration.Provider, m *RecoverInvalidModel) *RecoverInvalid {
	return &RecoverInvalid{c: c, m: m}
}

func (t *RecoverInvalid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *RecoverInvalid) EmailSubject() (string, error) {
	return loadTextTemplate(filepath.Join(t.c.CourierTemplatesRoot(), "recover/invalid/email.subject.gotmpl"), t.m)
}

func (t *RecoverInvalid) EmailBody() (string, error) {
	return loadTextTemplate(filepath.Join(t.c.CourierTemplatesRoot(), "recover/invalid/email.body.gotmpl"), t.m)
}
