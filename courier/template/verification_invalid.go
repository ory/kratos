package template

import (
	"encoding/json"
	"os"
)

type (
	VerificationInvalid struct {
		c TemplateConfig
		m *VerificationInvalidModel
	}
	VerificationInvalidModel struct {
		To string
	}
)

func NewVerificationInvalid(c TemplateConfig, m *VerificationInvalidModel) *VerificationInvalid {
	return &VerificationInvalid{c: c, m: m}
}

func (t *VerificationInvalid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *VerificationInvalid) EmailSubject() (string, error) {
	return LoadTextTemplate(os.DirFS(t.c.CourierTemplatesRoot()), "verification/invalid/email.subject.gotmpl", "verification/invalid/email.subject*", t.m,
		t.c.CourierTemplatesVerificationInvalid().Subject,
		t.c.CourierTemplatesVerificationInvalid().TemplateRoot,
	)
}

func (t *VerificationInvalid) EmailBody() (string, error) {
	return LoadHTMLTemplate(os.DirFS(t.c.CourierTemplatesRoot()), "verification/invalid/email.body.gotmpl", "verification/invalid/email.body*", t.m,
		t.c.CourierTemplatesVerificationInvalid().Body.HTML,
		t.c.CourierTemplatesVerificationInvalid().TemplateRoot,
	)
}

func (t *VerificationInvalid) EmailBodyPlaintext() (string, error) {
	return LoadTextTemplate(os.DirFS(t.c.CourierTemplatesRoot()), "verification/invalid/email.body.plaintext.gotmpl", "verification/invalid/email.body.plaintext*", t.m,
		t.c.CourierTemplatesVerificationInvalid().Body.PlainText,
		t.c.CourierTemplatesVerificationInvalid().TemplateRoot,
	)
}

func (t *VerificationInvalid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
