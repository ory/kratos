// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

// FlowName is the flow name.
//
// The flow name can be one of:
// - 'login'
// - 'registration'
// - 'settings'
// - 'recovery'
// - 'verification'
//
// swagger:ignore
type FlowName string

const (
	LoginFlow        FlowName = "login"
	RegistrationFlow FlowName = "registration"
	SettingsFlow     FlowName = "settings"
	RecoveryFlow     FlowName = "recovery"
	VerificationFlow FlowName = "verification"
)

func (t Type) String() string {
	return string(t)
}
