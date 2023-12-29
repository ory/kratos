// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/internal"
)

func TestNewEmailTemplateFromMessage(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	ctx := context.Background()

	for tmplType, expectedTmpl := range map[template.TemplateType]courier.EmailTemplate{
		template.TypeRecoveryInvalid:         email.NewRecoveryInvalid(reg, &email.RecoveryInvalidModel{To: "foo"}),
		template.TypeRecoveryValid:           email.NewRecoveryValid(reg, &email.RecoveryValidModel{To: "bar", RecoveryURL: "http://foo.bar"}),
		template.TypeRecoveryCodeValid:       email.NewRecoveryCodeValid(reg, &email.RecoveryCodeValidModel{To: "bar", RecoveryCode: "12345678"}),
		template.TypeRecoveryCodeInvalid:     email.NewRecoveryCodeInvalid(reg, &email.RecoveryCodeInvalidModel{To: "bar"}),
		template.TypeVerificationInvalid:     email.NewVerificationInvalid(reg, &email.VerificationInvalidModel{To: "baz"}),
		template.TypeVerificationValid:       email.NewVerificationValid(reg, &email.VerificationValidModel{To: "faz", VerificationURL: "http://bar.foo"}),
		template.TypeVerificationCodeInvalid: email.NewVerificationCodeInvalid(reg, &email.VerificationCodeInvalidModel{To: "baz"}),
		template.TypeVerificationCodeValid:   email.NewVerificationCodeValid(reg, &email.VerificationCodeValidModel{To: "faz", VerificationURL: "http://bar.foo", VerificationCode: "123456678"}),
		template.TypeTestStub:                email.NewTestStub(reg, &email.TestStubModel{To: "far", Subject: "test subject", Body: "test body"}),
		template.TypeLoginCodeValid:          email.NewLoginCodeValid(reg, &email.LoginCodeValidModel{To: "far", LoginCode: "123456"}),
		template.TypeRegistrationCodeValid:   email.NewRegistrationCodeValid(reg, &email.RegistrationCodeValidModel{To: "far", RegistrationCode: "123456"}),
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
