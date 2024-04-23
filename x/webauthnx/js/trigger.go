// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package js

import "fmt"

// swagger:enum WebAuthnTriggers
type WebAuthnTriggers string

const (
	WebAuthnTriggersWebAuthnRegistration         WebAuthnTriggers = "oryWebAuthnRegistration"
	WebAuthnTriggersWebAuthnLogin                WebAuthnTriggers = "oryWebAuthnLogin"
	WebAuthnTriggersPasskeyLogin                 WebAuthnTriggers = "oryPasskeyLogin"
	WebAuthnTriggersPasskeyLoginAutocompleteInit WebAuthnTriggers = "oryPasskeyLoginAutocompleteInit"
	WebAuthnTriggersPasskeyRegistration          WebAuthnTriggers = "oryPasskeyRegistration"
	WebAuthnTriggersPasskeySettingsRegistration  WebAuthnTriggers = "oryPasskeySettingsRegistration"
)

func (r WebAuthnTriggers) String() string {
	return fmt.Sprintf("window.%s", string(r))
}
