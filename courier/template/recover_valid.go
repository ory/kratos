package template

import (
	"path/filepath"

	"github.com/ory/kratos/driver/configuration"
)

type (
	RecoverValid struct {
		c configuration.Provider
		m *RecoverValidModel
	}
	RecoverValidModel struct {
		To string
	}
)

func NewRecoverValid(c configuration.Provider, m *RecoverValidModel) *RecoverValid {
	return &RecoverValid{c: c, m: m}
}

func (t *RecoverValid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *RecoverValid) EmailSubject() (string, error) {
	return loadTextTemplate(filepath.Join(t.c.CourierTemplatesRoot(), "recover/valid/email.subject.gotmpl"), t.m)
}

func (t *RecoverValid) EmailBody() (string, error) {
	return loadTextTemplate(filepath.Join(t.c.CourierTemplatesRoot(), "recover/valid/email.body.gotmpl"), t.m)
}
