package email

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ory/kratos/courier/template"
)

type (
	RecoveryValid struct {
		d template.Dependencies
		m *RecoveryValidModel
	}
	RecoveryValidModel struct {
		To          string
		RecoveryURL string
		Identity    map[string]interface{}
	}
)

func NewRecoveryValid(d template.Dependencies, m *RecoveryValidModel) *RecoveryValid {
	return &RecoveryValid{d: d, m: m}
}

func (t *RecoveryValid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *RecoveryValid) EmailSubject(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig(ctx).CourierTemplatesRoot()), "recovery/valid/email.subject.gotmpl", "recovery/valid/email.subject*", t.m, t.d.CourierConfig(ctx).CourierTemplatesRecoveryValid().Subject)
}

func (t *RecoveryValid) EmailBody(ctx context.Context) (string, error) {
	return template.LoadHTML(ctx, t.d, os.DirFS(t.d.CourierConfig(ctx).CourierTemplatesRoot()), "recovery/valid/email.body.gotmpl", "recovery/valid/email.body*", t.m, t.d.CourierConfig(ctx).CourierTemplatesRecoveryValid().Body.HTML)
}

func (t *RecoveryValid) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig(ctx).CourierTemplatesRoot()), "recovery/valid/email.body.plaintext.gotmpl", "recovery/valid/email.body.plaintext*", t.m, t.d.CourierConfig(ctx).CourierTemplatesRecoveryValid().Body.PlainText)
}

func (t *RecoveryValid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
