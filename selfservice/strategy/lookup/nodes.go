// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package lookup

import (
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
)

func NewRevealLookupNode() *node.Node {
	return node.NewInputField(node.LookupReveal, "true", node.LookupGroup, node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewInfoSelfServiceSettingsRevealLookup())
}

func NewRegenerateLookupNode() *node.Node {
	return node.NewInputField(
		node.LookupRegenerate, "true", node.LookupGroup, node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewInfoSelfServiceSettingsRegenerateLookup())
}

func NewDisableLookupNode() *node.Node {
	return node.NewInputField(node.LookupDisable, "true", node.LookupGroup, node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewInfoSelfServiceSettingsDisableLookup())
}

func NewConfirmLookupNode() *node.Node {
	return node.NewInputField(node.LookupConfirm, "true", node.LookupGroup, node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewInfoSelfServiceSettingsLookupConfirm())
}
