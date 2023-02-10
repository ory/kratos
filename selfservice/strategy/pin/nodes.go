package pin

import (
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
)

func NewPinNode(name string) *node.Node {
	return node.NewInputField(name, nil, node.PinGroup,
		node.InputAttributeTypePassword,
		node.WithRequiredInputAttribute).
		WithMetaLabel(text.NewInfoNodeInputPin())
}
