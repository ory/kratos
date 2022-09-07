package email

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/ory/kratos/courier/template"
)

type (
	VerificationValidOTP struct {
		d template.Dependencies
		m *VerificationValidOTPModel
	}
	VerificationValidOTPModel struct {
		To       string
		Code     string
		Identity map[string]interface{}
	}
)

func NewVerificationValidOTP(d template.Dependencies, m *VerificationValidOTPModel) *VerificationValidOTP {
	return &VerificationValidOTP{d: d, m: m}
}

func (t *VerificationValidOTP) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *VerificationValidOTP) EmailSubject(ctx context.Context) (string, error) {
	templatesDir := os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx))
	subject := t.d.CourierConfig().CourierTemplatesVerificationValid(ctx).Subject

	subject, err := template.LoadText(ctx, t.d, templatesDir, "otp/verification/valid/email.subject.gotmpl", "otp/verification/valid/email.subject*", t.m, subject)

	return strings.TrimSpace(subject), err
}

func (t *VerificationValidOTP) EmailBody(ctx context.Context) (string, error) {
	templatesDir := os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx))
	body := t.d.CourierConfig().CourierTemplatesVerificationValid(ctx).Body.HTML

	return template.LoadHTML(ctx, t.d, templatesDir, "otp/verification/valid/email.body.gotmpl", "otp/verification/valid/email.body*", t.m, body)
}

func (t *VerificationValidOTP) EmailBodyPlaintext(ctx context.Context) (string, error) {
	templatesDir := os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx))
	plaintextBody := t.d.CourierConfig().CourierTemplatesVerificationValid(ctx).Body.PlainText

	return template.LoadText(ctx, t.d, templatesDir, "otp/verification/valid/email.body.plaintext.gotmpl", "otp/verification/valid/email.body.plaintext*", t.m, plaintextBody)
}

func (t *VerificationValidOTP) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
