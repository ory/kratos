package sms

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ory/kratos/courier/template"
)

const verificationMsgBodyPath = "otp/verification/valid/otp.body.gotmpl"

type (
	VerificationMessage struct {
		d template.Dependencies
		m *VerificationMessageModel
	}

	VerificationMessageModel struct {
		To       string
		Code     string
		Identity map[string]interface{}
	}
)

func NewVerificationOTPMessage(d template.Dependencies, m *VerificationMessageModel) *VerificationMessage {
	return &VerificationMessage{d: d, m: m}
}

func (t *VerificationMessage) PhoneNumber() (string, error) {
	return t.m.To, nil
}

func (t *VerificationMessage) SMSBody(ctx context.Context) (string, error) {
	templatesDir := os.DirFS(t.d.CourierConfig(ctx).CourierTemplatesRoot())
	return template.LoadText(ctx, t.d, templatesDir, verificationMsgBodyPath, "otp/sms.body*", t.m, "")
}

func (t *VerificationMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
