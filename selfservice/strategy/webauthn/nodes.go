package webauthn

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
)

//go:embed js/register.js
var jsRegister []byte

//go:embed js/login.js
var jsLogin []byte

func NewWebAuthnConnectionTrigger(options string) *node.Node {
	return node.NewInputField(node.WebAuthnRegisterTrigger, "", node.WebAuthnGroup,
		node.InputAttributeTypeButton, node.WithInputAttributes(func(a *node.InputAttributes) {
			a.OnClick = strings.Replace(string(jsRegister), "injectWebAuthnOptions", options, 1)
		})).
		WithMetaLabel(text.NewInfoSelfServiceRegisterWebAuthn())
}

func NewWebAuthnConnectionInput() *node.Node {
	return node.NewInputField(node.WebAuthnRegister, "", node.WebAuthnGroup,
		node.InputAttributeTypeHidden)
}

func NewWebAuthnLoginTrigger(options string) *node.Node {
	return node.NewInputField(node.WebAuthnLoginTrigger, "", node.WebAuthnGroup,
		node.InputAttributeTypeButton, node.WithInputAttributes(func(a *node.InputAttributes) {
			a.OnClick = strings.Replace(string(jsLogin), "injectWebAuthnOptions", options, 1)
		})).
		WithMetaLabel(text.NewInfoSelfServiceLoginWebAuthn())
}

func NewWebAuthnLoginInput() *node.Node {
	return node.NewInputField(node.WebAuthnLogin, "", node.WebAuthnGroup,
		node.InputAttributeTypeHidden)
}

func NewWebAuthnConnectionName() *node.Node {
	return node.NewInputField(node.WebAuthnRegisterDisplayName, "", node.WebAuthnGroup, node.InputAttributeTypeText).
		WithMetaLabel(text.NewInfoSelfServiceRegisterWebAuthnDisplayName())
}

func NewWebAuthnUnlink(c *Credential) *node.Node {
	return node.NewInputField(node.WebAuthnRemove, fmt.Sprintf("%x", c.ID), node.WebAuthnGroup,
		node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewInfoSelfServiceRemoveWebAuthn(c.DisplayName, c.AddedAt))
}
