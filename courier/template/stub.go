package template

import (
	"encoding/json"
	"os"
)

type TestStub struct {
	c TemplateConfig
	m *TestStubModel
}

type TestStubModel struct {
	To      string
	Subject string
	Body    string
}

func NewTestStub(c TemplateConfig, m *TestStubModel) *TestStub {
	return &TestStub{c: c, m: m}
}

func (t *TestStub) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *TestStub) EmailSubject() (string, error) {
	return LoadTextTemplate(os.DirFS(t.c.CourierTemplatesRoot()), "test_stub/email.subject.gotmpl", "test_stub/email.subject*", t.m)
}

func (t *TestStub) EmailBody() (string, error) {
	return LoadHTMLTemplate(os.DirFS(t.c.CourierTemplatesRoot()), "test_stub/email.body.gotmpl", "test_stub/email.body*", t.m)
}

func (t *TestStub) EmailBodyPlaintext() (string, error) {
	return LoadTextTemplate(os.DirFS(t.c.CourierTemplatesRoot()), "test_stub/email.body.plaintext.gotmpl", "test_stub/email.body.plaintext*", t.m)
}

func (t *TestStub) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
