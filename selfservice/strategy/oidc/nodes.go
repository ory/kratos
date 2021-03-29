package oidc

import "github.com/ory/kratos/ui/node"

func NewLinkNode(provider string) *node.Node {
	return node.NewInputField("link", provider, node.OpenIDConnectGroup, node.InputAttributeTypeSubmit)
}

func NewUnlinkNode(provider string) *node.Node {
	return node.NewInputField("unlink", provider, node.OpenIDConnectGroup, node.InputAttributeTypeSubmit)
}
