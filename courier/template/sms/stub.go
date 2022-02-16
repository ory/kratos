package sms

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
		To       string
		Body     string
		Identity map[string]interface{}
	}
)

func NewTestStub(d template.Dependencies, m *TestStubModel) *TestStub {
	return &TestStub{d: d, m: m}
}

func (t *TestStub) PhoneNumber() (string, error) {
	return t.m.To, nil
}

func (t *TestStub) SMSBody(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig(ctx).CourierTemplatesRoot()), "otp/test_stub/sms.body.gotmpl", "otp/test_stub/sms.body*", t.m, "")
}

func (t *TestStub) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
