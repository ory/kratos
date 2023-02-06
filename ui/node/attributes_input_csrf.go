// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
