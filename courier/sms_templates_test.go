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
	"github.com/ory/kratos/courier/template/sms"
	"github.com/ory/kratos/internal"
)

func TestSMSTemplateType(t *testing.T) {
	for expectedType, tmpl := range map[courier.TemplateType]courier.SMSTemplate{
		courier.TypeOTP:      &sms.OTPMessage{},
		courier.TypeTestStub: &sms.TestStub{},
	} {
		t.Run(fmt.Sprintf("case=%s", expectedType), func(t *testing.T) {
			actualType, err := courier.SMSTemplateType(tmpl)
			require.NoError(t, err)
			require.Equal(t, expectedType, actualType)
		})
	}
}

func TestNewSMSTemplateFromMessage(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	ctx := context.Background()

	for tmplType, expectedTmpl := range map[courier.TemplateType]courier.SMSTemplate{
		courier.TypeOTP:      sms.NewOTPMessage(reg, &sms.OTPMessageModel{To: "+12345678901"}),
		courier.TypeTestStub: sms.NewTestStub(reg, &sms.TestStubModel{To: "+12345678901", Body: "test body"}),
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
