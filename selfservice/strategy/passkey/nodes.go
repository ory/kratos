// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey

import (
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x/webauthnx/js"
)

func injectOptions(o []byte) *node.Node {
	return &node.Node{
		Type:  node.Input,
		Group: node.PasskeyGroup,
		Meta:  &node.Meta{},
		Attributes: &node.InputAttributes{
			Name:       node.PasskeyCreateData,
			Type:       node.InputAttributeTypeHidden,
			FieldValue: string(o),
		},
	}
}

func passkeyRegister() *node.Node {
	return &node.Node{
		Type:  node.Input,
		Group: node.PasskeyGroup,
		Meta:  &node.Meta{},
		Attributes: &node.InputAttributes{
			Name: node.PasskeyRegister,
			Type: node.InputAttributeTypeHidden,
		},
	}
}

func passkeyRegisterTrigger() *node.Node {
	return &node.Node{
		Type:  node.Input,
		Group: node.PasskeyGroup,
		Meta:  &node.Meta{Label: text.NewInfoSelfServiceRegistrationRegisterPasskey()},
		Attributes: &node.InputAttributes{
			Name:           node.PasskeyRegisterTrigger,
			Type:           node.InputAttributeTypeButton,
			OnClick:        js.WebAuthnTriggersPasskeyRegistration.String() + "()", // defined in webauthn.js
			OnClickTrigger: js.WebAuthnTriggersPasskeyRegistration,
		},
	}
}
