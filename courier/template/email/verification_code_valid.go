package email

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/ory/kratos/courier/template"
)

type (
	VerificationCodeValid struct {
		d template.Dependencies
		m *VerificationCodeValidModel
	}
	VerificationCodeValidModel struct {
		To              string
		VerificationURL string
		Identity        map[string]interface{}
	}
)

func NewVerificationCodeValid(d template.Dependencies, m *VerificationCodeValidModel) *VerificationCodeValid {
	return &VerificationCodeValid{d: d, m: m}
}

func (t *VerificationCodeValid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *VerificationCodeValid) EmailSubject(ctx context.Context) (string, error) {
	subject, err := template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "verification/valid/email.subject.gotmpl", "verification/valid/email.subject*", t.m, t.d.CourierConfig().CourierTemplatesVerificationValid(ctx).Subject)

	return strings.TrimSpace(subject), err
}

func (t *VerificationCodeValid) EmailBody(ctx context.Context) (string, error) {
	return template.LoadHTML(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "verification/valid/email.body.gotmpl", "verification/valid/email.body*", t.m, t.d.CourierConfig().CourierTemplatesVerificationValid(ctx).Body.HTML)
}

func (t *VerificationCodeValid) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "verification/valid/email.body.plaintext.gotmpl", "verification/valid/email.body.plaintext*", t.m, t.d.CourierConfig().CourierTemplatesVerificationValid(ctx).Body.PlainText)
}

func (t *VerificationCodeValid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
