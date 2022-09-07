package email

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/ory/kratos/courier/template"
)

type (
	RecoveryValidOTP struct {
		d template.Dependencies
		m *RecoveryValidOTPModel
	}
	RecoveryValidOTPModel struct {
		To       string
		Code     string
		Identity map[string]interface{}
	}
)

func NewRecoveryValidOTP(d template.Dependencies, m *RecoveryValidOTPModel) *RecoveryValidOTP {
	return &RecoveryValidOTP{d: d, m: m}
}

func (t *RecoveryValidOTP) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *RecoveryValidOTP) EmailSubject(ctx context.Context) (string, error) {
	templatesDir := os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx))
	subject := t.d.CourierConfig().CourierTemplatesRecoveryValid(ctx).Subject

	subject, err := template.LoadText(ctx, t.d, templatesDir, "otp/recovery/valid/email.subject.gotmpl", "otp/recovery/valid/email.subject*", t.m, subject)

	return strings.TrimSpace(subject), err
}

func (t *RecoveryValidOTP) EmailBody(ctx context.Context) (string, error) {
	templatesDir := os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx))
	body := t.d.CourierConfig().CourierTemplatesRecoveryValid(ctx).Body.HTML

	return template.LoadHTML(ctx, t.d, templatesDir, "otp/recovery/valid/email.body.gotmpl", "otp/recovery/valid/email.body*", t.m, body)
}

func (t *RecoveryValidOTP) EmailBodyPlaintext(ctx context.Context) (string, error) {
	templatesDir := os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx))
	bodyPlaintext := t.d.CourierConfig().CourierTemplatesRecoveryValid(ctx).Body.PlainText

	return template.LoadText(ctx, t.d, templatesDir, "otp/recovery/valid/email.body.plaintext.gotmpl", "otp/recovery/valid/email.body.plaintext*", t.m, bodyPlaintext)
}

func (t *RecoveryValidOTP) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
