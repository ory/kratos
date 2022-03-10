package login

import (
	"context"

	"github.com/ory/kratos/ui/node"
)

var ctx = context.Background()

func sortNodes(n node.Nodes) error {
	return n.SortBySchema(ctx,
		node.SortByGroups([]node.Group{
			node.OpenIDConnectGroup,
			node.DefaultGroup,
			node.WebAuthnGroup,
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
