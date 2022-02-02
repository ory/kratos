package template

import (
	"encoding/json"
	"os"
)

type (
	RecoveryInvalid struct {
		c TemplateConfig
		m *RecoveryInvalidModel
	}
	RecoveryInvalidModel struct {
		To string
	}
)

func NewRecoveryInvalid(c TemplateConfig, m *RecoveryInvalidModel) *RecoveryInvalid {
	return &RecoveryInvalid{c: c, m: m}
}

func (t *RecoveryInvalid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *RecoveryInvalid) EmailSubject() (string, error) {
	return LoadTextTemplate(os.DirFS(t.c.CourierTemplatesRoot()), "recovery/invalid/email.subject.gotmpl", "recovery/invalid/email.subject*", t.m)
}

func (t *RecoveryInvalid) EmailBody() (string, error) {
	return LoadHTMLTemplate(os.DirFS(t.c.CourierTemplatesRoot()), "recovery/invalid/email.body.gotmpl", "recovery/invalid/email.body*", t.m)
}

func (t *RecoveryInvalid) EmailBodyPlaintext() (string, error) {
	return LoadTextTemplate(os.DirFS(t.c.CourierTemplatesRoot()), "recovery/invalid/email.body.plaintext.gotmpl", "recovery/invalid/email.body.plaintext*", t.m)
}

func (t *RecoveryInvalid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
