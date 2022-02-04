package template

import (
	"context"
	"encoding/json"
	"os"
)

type (
	VerificationInvalid struct {
		d TemplateDependencies
		m *VerificationInvalidModel
	}
	VerificationInvalidModel struct {
		To string
	}
)

func NewVerificationInvalid(d TemplateDependencies, m *VerificationInvalidModel) *VerificationInvalid {
	return &VerificationInvalid{d: d, m: m}
}

func (t *VerificationInvalid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *VerificationInvalid) EmailSubject(ctx context.Context) (string, error) {
	return LoadTextTemplate(ctx, t.d, os.DirFS(t.d.CourierConfig(ctx).CourierTemplatesRoot()), "verification/invalid/email.subject.gotmpl", "verification/invalid/email.subject*", t.m, t.d.CourierConfig(ctx).CourierTemplatesVerificationInvalid().Subject)
}

func (t *VerificationInvalid) EmailBody(ctx context.Context) (string, error) {
	return LoadHTMLTemplate(ctx, t.d, os.DirFS(t.d.CourierConfig(ctx).CourierTemplatesRoot()), "verification/invalid/email.body.gotmpl", "verification/invalid/email.body*", t.m, t.d.CourierConfig(ctx).CourierTemplatesVerificationInvalid().Body.HTML)
}

func (t *VerificationInvalid) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return LoadTextTemplate(ctx, t.d, os.DirFS(t.d.CourierConfig(ctx).CourierTemplatesRoot()), "verification/invalid/email.body.plaintext.gotmpl", "verification/invalid/email.body.plaintext*", t.m, t.d.CourierConfig(ctx).CourierTemplatesVerificationInvalid().Body.PlainText)
}

func (t *VerificationInvalid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
