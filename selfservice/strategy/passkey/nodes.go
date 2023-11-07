// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey

import (
	_ "embed"
	"fmt"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/stringsx"
)

func NewWebAuthnRegistrationTrigger(options string) *node.Node {
	return node.NewInputField(node.WebAuthnRegisterTrigger, "", node.WebAuthnGroup,
		node.InputAttributeTypeButton, node.WithInputAttributes(func(a *node.InputAttributes) {
			a.OnClick = "window.__oryWebAuthnRegistration(" + options + ")"
		}))
}

func NewWebAuthnConnectionInput() *node.Node {
	return node.NewInputField(node.WebAuthnRegister, "", node.PasskeyGroup,
		node.InputAttributeTypeHidden)
}

func NewWebAuthnLoginTrigger(options string) *node.Node {
	return node.NewInputField(node.WebAuthnLoginTrigger, "", node.WebAuthnGroup,
		node.InputAttributeTypeButton, node.WithInputAttributes(func(a *node.InputAttributes) {
			a.OnClick = "window.__oryWebAuthnLogin(" + options + ")"
		}))
}

func NewWebAuthnLoginInput() *node.Node {
	return node.NewInputField(node.WebAuthnLogin, "", node.WebAuthnGroup,
		node.InputAttributeTypeHidden)
}

func NewWebAuthnConnectionName() *node.Node {
	return node.NewInputField(node.WebAuthnRegisterDisplayName, "", node.WebAuthnGroup, node.InputAttributeTypeText).
		WithMetaLabel(text.NewInfoSelfServiceRegisterWebAuthnDisplayName())
}

func NewWebAuthnUnlink(c *identity.CredentialWebAuthn) *node.Node {
	return node.NewInputField(node.WebAuthnRemove, fmt.Sprintf("%x", c.ID), node.WebAuthnGroup,
		node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewInfoSelfServiceRemoveWebAuthn(stringsx.Coalesce(c.DisplayName, "unnamed"), c.AddedAt))
}
