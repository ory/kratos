package sms

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ory/kratos/courier/template"
)

type (
	OTPMessage struct {
		d template.Dependencies
		m *OTPMessageModel
	}

	OTPMessageModel struct {
		To       string
		Code     string
		Identity map[string]interface{}
	}
)

func NewOTPMessage(d template.Dependencies, m *OTPMessageModel) *OTPMessage {
	return &OTPMessage{d: d, m: m}
}

func (t *OTPMessage) PhoneNumber() (string, error) {
	return t.m.To, nil
}

func (t *OTPMessage) SMSBody(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig(ctx).CourierTemplatesRoot()), "otp/sms.body.gotmpl", "otp/sms.body*", t.m, "")
}

func (t *OTPMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
