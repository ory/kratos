package template

import (
	"path/filepath"

	"github.com/ory/kratos/driver/config"
)

type TestStub struct {
	c *config.Config
	m *TestStubModel
}

type TestStubModel struct {
	To      string
	Subject string
	Body    string
}

func NewTestStub(c *config.Config, m *TestStubModel) *TestStub {
	return &TestStub{c: c, m: m}
}

func (t *TestStub) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *TestStub) EmailSubject() (string, error) {
	return loadTextTemplate(filepath.Join(t.c.CourierTemplatesRoot(), "test_stub/email.subject.gotmpl"), t.m)
}

func (t *TestStub) EmailBody() (string, error) {
	return loadTextTemplate(filepath.Join(t.c.CourierTemplatesRoot(), "test_stub/email.body.gotmpl"), t.m)
}
