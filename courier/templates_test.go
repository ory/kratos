package courier_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/internal"
)

func TestGetTemplateType(t *testing.T) {
	for expectedType, tmpl := range map[courier.TemplateType]courier.EmailTemplate{
		courier.TypeRecoveryInvalid:     &template.RecoveryInvalid{},
		courier.TypeRecoveryValid:       &template.RecoveryValid{},
		courier.TypeVerificationInvalid: &template.VerificationInvalid{},
		courier.TypeVerificationValid:   &template.VerificationValid{},
		courier.TypeTestStub:            &template.TestStub{},
	} {
		t.Run(fmt.Sprintf("case=%s", expectedType), func(t *testing.T) {
			actualType, err := courier.GetTemplateType(tmpl)
			require.NoError(t, err)
			require.Equal(t, expectedType, actualType)

		})

	}
}

func TestNewEmailTemplateFromMessage(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	ctx := context.Background()

	for tmplType, expectedTmpl := range map[courier.TemplateType]courier.EmailTemplate{
		courier.TypeRecoveryInvalid:     template.NewRecoveryInvalid(reg, &template.RecoveryInvalidModel{To: "foo"}),
		courier.TypeRecoveryValid:       template.NewRecoveryValid(reg, &template.RecoveryValidModel{To: "bar", RecoveryURL: "http://foo.bar"}),
		courier.TypeVerificationInvalid: template.NewVerificationInvalid(reg, &template.VerificationInvalidModel{To: "baz"}),
		courier.TypeVerificationValid:   template.NewVerificationValid(reg, &template.VerificationValidModel{To: "faz", VerificationURL: "http://bar.foo"}),
		courier.TypeTestStub:            template.NewTestStub(reg, &template.TestStubModel{To: "far", Subject: "test subject", Body: "test body"}),
	} {
		t.Run(fmt.Sprintf("case=%s", tmplType), func(t *testing.T) {
			tmplData, err := json.Marshal(expectedTmpl)
			require.NoError(t, err)

			m := courier.Message{TemplateType: tmplType, TemplateData: tmplData}
			actualTmpl, err := courier.NewEmailTemplateFromMessage(reg, m)
			require.NoError(t, err)

			require.IsType(t, expectedTmpl, actualTmpl)

			expectedRecipient, err := expectedTmpl.EmailRecipient()
			require.NoError(t, err)
			actualRecipient, err := actualTmpl.EmailRecipient()
			require.NoError(t, err)
			require.Equal(t, expectedRecipient, actualRecipient)

			expectedSubject, err := expectedTmpl.EmailSubject(ctx)
			require.NoError(t, err)
			actualSubject, err := actualTmpl.EmailSubject(ctx)
			require.NoError(t, err)
			require.Equal(t, expectedSubject, actualSubject)

			expectedBody, err := expectedTmpl.EmailBody(ctx)
			require.NoError(t, err)
			actualBody, err := actualTmpl.EmailBody(ctx)
			require.NoError(t, err)
			require.Equal(t, expectedBody, actualBody)

			expectedBodyPlaintext, err := expectedTmpl.EmailBodyPlaintext(ctx)
			require.NoError(t, err)
			actualBodyPlaintext, err := actualTmpl.EmailBodyPlaintext(ctx)
			require.NoError(t, err)
			require.Equal(t, expectedBodyPlaintext, actualBodyPlaintext)

		})
	}
}
