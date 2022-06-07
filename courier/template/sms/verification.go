package sms

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ory/kratos/courier/template"
)

type (
	VerificationMessage struct {
		d template.Dependencies
		m *VerificationMessageModel
	}
	VerificationMessageModel struct {
		To              string
		VerificationURL string
		Identity        map[string]interface{}
	}
)

func NewVerificationMessage(d template.Dependencies, m *VerificationMessageModel) *VerificationMessage {
	return &VerificationMessage{d: d, m: m}
}

func (t *VerificationMessage) PhoneNumber() (string, error) {
	return t.m.To, nil
}

func (t *VerificationMessage) SMSBody(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig(ctx).CourierTemplatesRoot()), "verification/valid/sms.body.gotmpl", "verification/valid/sms.body*", t.m, t.d.CourierConfig(ctx).CourierTemplatesVerificationValid().Body.PlainText)
}

func (t *VerificationMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
