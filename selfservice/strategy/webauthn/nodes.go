package webauthn

import (
	"crypto/sha512"
	_ "embed"
	"encoding/base64"
	"fmt"

	"github.com/ory/x/stringsx"

	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
)

func NewWebAuthnConnectionTrigger(options string) *node.Node {
	return node.NewInputField(node.WebAuthnRegisterTrigger, "", node.WebAuthnGroup,
		node.InputAttributeTypeButton, node.WithInputAttributes(func(a *node.InputAttributes) {
			a.OnClick = "window.__oryWebAuthnRegistration(" + options + ")"
		}))
}

func NewWebAuthnScript(src string, contents []byte) *node.Node {
	integrity := sha512.Sum512(contents)
	return node.NewScriptField(node.WebAuthnScript, src, node.WebAuthnGroup, fmt.Sprintf("sha512-%s", base64.StdEncoding.EncodeToString(integrity[:])))
}

func NewWebAuthnConnectionInput() *node.Node {
	return node.NewInputField(node.WebAuthnRegister, "", node.WebAuthnGroup,
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

func NewWebAuthnUnlink(c *Credential) *node.Node {
	return node.NewInputField(node.WebAuthnRemove, fmt.Sprintf("%x", c.ID), node.WebAuthnGroup,
		node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewInfoSelfServiceRemoveWebAuthn(stringsx.Coalesce(c.DisplayName, "unnamed"), c.AddedAt))
}
