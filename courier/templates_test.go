package courier_test

import (
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
	conf := internal.NewConfigurationWithDefaults()
	for tmplType, expectedTmpl := range map[courier.TemplateType]courier.EmailTemplate{
		courier.TypeRecoveryInvalid:     template.NewRecoveryInvalid(conf, &template.RecoveryInvalidModel{To: "foo"}),
		courier.TypeRecoveryValid:       template.NewRecoveryValid(conf, &template.RecoveryValidModel{To: "bar", RecoveryURL: "http://foo.bar"}),
		courier.TypeVerificationInvalid: template.NewVerificationInvalid(conf, &template.VerificationInvalidModel{To: "baz"}),
		courier.TypeVerificationValid:   template.NewVerificationValid(conf, &template.VerificationValidModel{To: "faz", VerificationURL: "http://bar.foo"}),
		courier.TypeTestStub:            template.NewTestStub(conf, &template.TestStubModel{To: "far", Subject: "test subject", Body: "test body"}),
	} {
		t.Run(fmt.Sprintf("case=%s", tmplType), func(t *testing.T) {
			tmplData, err := json.Marshal(expectedTmpl)
			require.NoError(t, err)

			m := courier.Message{TemplateType: tmplType, TemplateData: tmplData}
			actualTmpl, err := courier.NewEmailTemplateFromMessage(conf, m)
			require.NoError(t, err)

			require.IsType(t, expectedTmpl, actualTmpl)

			expectedRecipient, err := expectedTmpl.EmailRecipient()
			require.NoError(t, err)
			actualRecipient, err := actualTmpl.EmailRecipient()
			require.NoError(t, err)
			require.Equal(t, expectedRecipient, actualRecipient)

			expectedSubject, err := expectedTmpl.EmailSubject()
			require.NoError(t, err)
			actualSubject, err := actualTmpl.EmailSubject()
			require.NoError(t, err)
			require.Equal(t, expectedSubject, actualSubject)

			expectedBody, err := expectedTmpl.EmailBody()
			require.NoError(t, err)
			actualBody, err := actualTmpl.EmailBody()
			require.NoError(t, err)
			require.Equal(t, expectedBody, actualBody)

			expectedBodyPlaintext, err := expectedTmpl.EmailBodyPlaintext()
			require.NoError(t, err)
			actualBodyPlaintext, err := actualTmpl.EmailBodyPlaintext()
			require.NoError(t, err)
			require.Equal(t, expectedBodyPlaintext, actualBodyPlaintext)

		})
	}
}
