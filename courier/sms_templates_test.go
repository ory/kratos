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
	"github.com/ory/kratos/courier/template/sms"
	"github.com/ory/kratos/pkg"
)

func TestSMSTemplateType(t *testing.T) {
	for expectedType, tmpl := range map[template.TemplateType]courier.SMSTemplate{
		template.TypeVerificationCodeValid: &sms.VerificationCodeValid{},
		template.TypeTestStub:              &sms.TestStub{},
	} {
		t.Run(fmt.Sprintf("case=%s", expectedType), func(t *testing.T) {
			require.Equal(t, expectedType, tmpl.TemplateType())
		})
	}
}

func TestNewSMSTemplateFromMessage(t *testing.T) {
	_, reg := pkg.NewFastRegistryWithMocks(t)
	ctx := context.Background()

	for tmplType, expectedTmpl := range map[template.TemplateType]courier.SMSTemplate{
		template.TypeVerificationCodeValid:    sms.NewVerificationCodeValid(reg, &sms.VerificationCodeValidModel{To: "+12345678901"}),
		template.TypeTestStub:                 sms.NewTestStub(&sms.TestStubModel{To: "+12345678901", Body: "test body"}),
		template.TypeVerifiableAddressChanged: sms.NewVerifiableAddressChanged(reg, &sms.VerifiableAddressChangedModel{To: "+12345678901", ChangedAt: "2026-04-21T12:00:00Z", Identity: map[string]any{"ID": "00000000-0000-0000-0000-000000000001"}}),
		template.TypeAuthenticatorKeyAdded:    sms.NewAuthenticatorKeyAdded(reg, &sms.AuthenticatorKeyAddedModel{To: "+12345678901", AddedAt: "2026-04-21T12:00:00Z", Identity: map[string]any{"ID": "00000000-0000-0000-0000-000000000001"}}),
	} {
		t.Run(fmt.Sprintf("case=%s", tmplType), func(t *testing.T) {
			tmplData, err := json.Marshal(expectedTmpl)
			require.NoError(t, err)

			m := courier.Message{TemplateType: tmplType, TemplateData: tmplData}
			actualTmpl, err := courier.NewSMSTemplateFromMessage(reg, m)
			require.NoError(t, err)

			require.IsType(t, expectedTmpl, actualTmpl)

			expectedRecipient, err := expectedTmpl.PhoneNumber()
			require.NoError(t, err)
			actualRecipient, err := actualTmpl.PhoneNumber()
			require.NoError(t, err)
			require.Equal(t, expectedRecipient, actualRecipient)

			expectedBody, err := expectedTmpl.SMSBody(ctx)
			require.NoError(t, err)
			actualBody, err := actualTmpl.SMSBody(ctx)
			require.NoError(t, err)
			require.Equal(t, expectedBody, actualBody)
		})
	}
}
