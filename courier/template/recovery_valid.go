package template

import (
	"encoding/json"

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
		Identity    map[string]interface{}
	}
)

func NewRecoveryValid(c *config.Config, m *RecoveryValidModel) *RecoveryValid {
	return &RecoveryValid{c: c, m: m}
}

func (t *RecoveryValid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *RecoveryValid) EmailSubject() (string, error) {
	return loadTextTemplate(t.c.CourierTemplatesRoot(), "recovery/valid/email.subject.gotmpl", "recovery/valid/email.subject*", t.m)
}

func (t *RecoveryValid) EmailBody() (string, error) {
	return loadHTMLTemplate(t.c.CourierTemplatesRoot(), "recovery/valid/email.body.gotmpl", "recovery/valid/email.body*", t.m)
}

func (t *RecoveryValid) EmailBodyPlaintext() (string, error) {
	return loadTextTemplate(t.c.CourierTemplatesRoot(), "recovery/valid/email.body.plaintext.gotmpl", "recovery/valid/email.body.plaintext*", t.m)
}

func (t *RecoveryValid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
