package template

import (
	"encoding/json"

	"github.com/ory/kratos/driver/config"
)

type (
	VerificationInvalid struct {
		c *config.Config
		m *VerificationInvalidModel
	}
	VerificationInvalidModel struct {
		To string
	}
)

func NewVerificationInvalid(c *config.Config, m *VerificationInvalidModel) *VerificationInvalid {
	return &VerificationInvalid{c: c, m: m}
}

func (t *VerificationInvalid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *VerificationInvalid) EmailSubject() (string, error) {
	return loadTextTemplate(t.c.CourierTemplatesRoot(), "verification/invalid/email.subject.gotmpl", "verification/invalid/email.subject*", t.m)
}

func (t *VerificationInvalid) EmailBody() (string, error) {
	return loadHTMLTemplate(t.c.CourierTemplatesRoot(), "verification/invalid/email.body.gotmpl", "verification/invalid/email.body*", t.m)
}

func (t *VerificationInvalid) EmailBodyPlaintext() (string, error) {
	return loadTextTemplate(t.c.CourierTemplatesRoot(), "verification/invalid/email.body.plaintext.gotmpl", "verification/invalid/email.body.plaintext*", t.m)
}

func (t *VerificationInvalid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
