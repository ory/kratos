// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"context"

	"github.com/ory/kratos/ui/node"
)

func SortNodes(ctx context.Context, n node.Nodes, schemaRef string) error {
	return n.SortBySchema(ctx,
		node.SortBySchema(schemaRef),
		node.SortByGroups([]node.UiNodeGroup{
			node.OpenIDConnectGroup,
			node.SAMLGroup,
			node.DefaultGroup,
			node.WebAuthnGroup,
			node.PasskeyGroup,
			node.CodeGroup,
			node.PasswordGroup,
			node.CaptchaGroup,
			node.ProfileGroup,
		}),
		node.SortUpdateOrder(node.PasswordLoginOrder),
		node.SortUseOrderAppend([]string{
			// WebAuthn
			node.WebAuthnRemove,
			node.WebAuthnRegisterDisplayName,
			node.WebAuthnRegister,
		}),
	)
}
