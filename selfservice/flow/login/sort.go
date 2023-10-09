// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"

	"github.com/ory/kratos/ui/node"
)

func sortNodes(ctx context.Context, n node.Nodes) error {
	return n.SortBySchema(ctx,
		node.SortByGroups([]node.UiNodeGroup{
			node.OpenIDConnectGroup,
			node.DefaultGroup,
			node.WebAuthnGroup,
			node.CodeGroup,
			node.PasswordGroup,
			node.TOTPGroup,
			node.LookupGroup,
		}),
		node.SortUseOrder([]string{
			"csrf_token",
			"identifier",
			"password",
		}),
	)
}
