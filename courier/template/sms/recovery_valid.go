package sms

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ory/kratos/courier/template"
)

const recoveryMsgBodyPath = "otp/recovery/valid/sms.body.gotmpl"

type (
	RecoveryMessage struct {
		d template.Dependencies
		m *RecoveryMessageModel
	}

	RecoveryMessageModel struct {
		To       string
		Code     string
		Identity map[string]interface{}
	}
)

func NewRecoveryOTPMessage(d template.Dependencies, m *RecoveryMessageModel) *RecoveryMessage {
	return &RecoveryMessage{d: d, m: m}
}

func (t *RecoveryMessage) PhoneNumber() (string, error) {
	return t.m.To, nil
}

func (t *RecoveryMessage) SMSBody(ctx context.Context) (string, error) {
	templatesDir := os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx))
	return template.LoadText(ctx, t.d, templatesDir, recoveryMsgBodyPath, "otp/sms.body*", t.m, "")
}

func (t *RecoveryMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
