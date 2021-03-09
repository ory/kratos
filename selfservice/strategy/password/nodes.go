package password

import "github.com/ory/kratos/ui/node"

func NewPasswordNode(name string) *node.Node {
	return node.NewInputField(name, nil, node.PasswordGroup, node.InputAttributeTypePassword, node.WithRequiredInputAttribute)
}
