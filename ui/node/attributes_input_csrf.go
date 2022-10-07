// Copyright © 2022 Ory Corp

package node

import "github.com/ory/kratos/x"

func NewCSRFNode(token string) *Node {
	return &Node{
		Type:  Input,
		Group: DefaultGroup,
		Attributes: &InputAttributes{
			Name:       x.CSRFTokenName,
			Type:       InputAttributeTypeHidden,
			FieldValue: token,
			Required:   true,
		},
	}
}
