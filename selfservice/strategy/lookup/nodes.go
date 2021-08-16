package lookup

import (
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
)

func NewRevealLookupNode() *node.Node {
	return node.NewInputField(node.LookupReveal, "true", node.LookupGroup,
		node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewInfoSelfServiceSettingsRevealLookup())
}

func NewRegenerateLookupNode() *node.Node {
	return node.NewInputField(node.LookupRegenerate, "true", node.LookupGroup,
		node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewInfoSelfServiceSettingsRegenerateLookup())
}

func NewConfirmLookupNode() *node.Node {
	return node.NewInputField(node.LookupConfirm, "true", node.LookupGroup,
		node.InputAttributeTypeSubmit).
		WithMetaLabel(text.NewInfoSelfServiceSettingsLookupConfirm())
}

func NewLookupNode(secrets []string) *node.Node {
	return node.NewTextField(node.LookupCodes, text.NewInfoSelfServiceSettingsLookupSecrets(secrets), node.LookupGroup).WithMetaLabel(text.NewInfoSelfServiceSettingsLookupSecretsLabel())
}
