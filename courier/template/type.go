// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package template

// A Template's type
//
// swagger:enum TemplateType
type TemplateType string

const (
	TypeRecoveryInvalid         TemplateType = "recovery_invalid"
	TypeRecoveryValid           TemplateType = "recovery_valid"
	TypeRecoveryCodeInvalid     TemplateType = "recovery_code_invalid"
	TypeRecoveryCodeValid       TemplateType = "recovery_code_valid"
	TypeVerificationInvalid     TemplateType = "verification_invalid"
	TypeVerificationValid       TemplateType = "verification_valid"
	TypeVerificationCodeInvalid TemplateType = "verification_code_invalid"
	TypeVerificationCodeValid   TemplateType = "verification_code_valid"
	TypeTestStub                TemplateType = "stub"
	TypeLoginCodeValid          TemplateType = "login_code_valid"
	TypeRegistrationCodeValid   TemplateType = "registration_code_valid"
)
