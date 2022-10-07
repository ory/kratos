// Copyright © 2022 Ory Corp

package password

import (
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
)

func NewPasswordNode(name string, autocomplete node.UiNodeInputAttributeAutocomplete) *node.Node {
	return node.NewInputField(name, nil, node.PasswordGroup,
		node.InputAttributeTypePassword,
		node.WithRequiredInputAttribute,
		node.WithInputAttributes(func(a *node.InputAttributes) {
			a.Autocomplete = autocomplete
		})).
		WithMetaLabel(text.NewInfoNodeInputPassword())
}
