package password

import "github.com/ory/kratos/ui/node"

func NewPasswordNode() *node.Node {
	return node.NewInputField("password", nil,node.PasswordGroup, node.InputAttributeTypePassword,  node.WithRequiredInputAttribute)
}
