package templates

import (
	"path/filepath"

	"github.com/ory/kratos/driver/configuration"
)

type (
	VerifyValid struct {
		c configuration.Provider
		m *VerifyValidModel
	}
	VerifyValidModel struct {
		To        string
		VerifyURL string
	}
)

func NewVerifyValid(c configuration.Provider, m *VerifyValidModel) *VerifyValid {
	return &VerifyValid{c: c, m: m}
}

func (t *VerifyValid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *VerifyValid) EmailSubject() (string, error) {
	return loadTextTemplate(filepath.Join(t.c.CourierTemplatesRoot(), "verify/valid/email.subject.gotmpl"), t.m)
}

func (t *VerifyValid) EmailBody() (string, error) {
	return loadTextTemplate(filepath.Join(t.c.CourierTemplatesRoot(), "verify/valid/email.body.gotmpl"), t.m)
}
