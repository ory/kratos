package login

import (
	"context"

	"github.com/ory/kratos/ui/node"
)

func sortNodes(ctx context.Context, n node.Nodes) error {
	return n.SortBySchema(ctx,
		node.SortByGroups([]node.Group{
			node.DefaultGroup,
			node.WebAuthnGroup,
			node.OpenIDConnectGroup,
			node.PasswordGroup,
			node.TOTPGroup,
			node.LookupGroup,
		}),
		node.SortUseOrder([]string{
			"password_identifier",
			"password",
		}),
	)
}
