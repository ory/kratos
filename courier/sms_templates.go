// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/courier/template/sms"
)

type SMSTemplate interface {
	json.Marshaler
	SMSBody(context.Context) (string, error)
	PhoneNumber() (string, error)
	TemplateType() template.TemplateType
}

func NewSMSTemplateFromMessage(d template.Dependencies, m Message) (SMSTemplate, error) {
	switch m.TemplateType {
	case template.TypeVerificationCodeValid:
		var t sms.VerificationCodeValidModel
		if err := json.Unmarshal(m.TemplateData, &t); err != nil {
			return nil, err
		}
		return sms.NewVerificationCodeValid(d, &t), nil
	case template.TypeTestStub:
		var t sms.TestStubModel
		if err := json.Unmarshal(m.TemplateData, &t); err != nil {
			return nil, err
		}
		return sms.NewTestStub(d, &t), nil
	case template.TypeLoginCodeValid:
		var t sms.LoginCodeValidModel
		if err := json.Unmarshal(m.TemplateData, &t); err != nil {
			return nil, err
		}
		return sms.NewLoginCodeValid(d, &t), nil
	case template.TypeRegistrationCodeValid:
		var t sms.RegistrationCodeValidModel
		if err := json.Unmarshal(m.TemplateData, &t); err != nil {
			return nil, err
		}
		return sms.NewRegistrationCodeValid(d, &t), nil

	default:
		return nil, errors.Errorf("received unexpected message template type: %s", m.TemplateType)
	}
}
