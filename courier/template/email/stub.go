package email

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ory/kratos/courier/template"
)

type (
	TestStub struct {
		d template.Dependencies
		m *TestStubModel
	}
	TestStubModel struct {
		To      string
		Subject string
		Body    string
	}
)

func NewTestStub(d template.Dependencies, m *TestStubModel) *TestStub {
	return &TestStub{d: d, m: m}
}

func (t *TestStub) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *TestStub) EmailSubject(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig(ctx).CourierTemplatesRoot()), "test_stub/email.subject.gotmpl", "test_stub/email.subject*", t.m, "")
}

func (t *TestStub) EmailBody(ctx context.Context) (string, error) {
	return template.LoadHTML(ctx, t.d, os.DirFS(t.d.CourierConfig(ctx).CourierTemplatesRoot()), "test_stub/email.body.gotmpl", "test_stub/email.body*", t.m, "")
}

func (t *TestStub) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig(ctx).CourierTemplatesRoot()), "test_stub/email.body.plaintext.gotmpl", "test_stub/email.body.plaintext*", t.m, "")
}

func (t *TestStub) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
