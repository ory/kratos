package template

import (
	"path/filepath"

	"github.com/ory/kratos/driver/config"
)

type (
	RecoveryValid struct {
		c *config.Config
		m *RecoveryValidModel
	}
	RecoveryValidModel struct {
		To          string
		RecoveryURL string
	}
)

func NewRecoveryValid(c *config.Config, m *RecoveryValidModel) *RecoveryValid {
	return &RecoveryValid{c: c, m: m}
}

func (t *RecoveryValid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *RecoveryValid) EmailSubject() (string, error) {
	return loadTextTemplate(filepath.Join(t.c.CourierTemplatesRoot(), "recovery/valid/email.subject.gotmpl"), t.m)
}

func (t *RecoveryValid) EmailBody() (string, error) {
	return loadTextTemplate(filepath.Join(t.c.CourierTemplatesRoot(), "recovery/valid/email.body.gotmpl"), t.m)
}
