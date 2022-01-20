package login

import (
	"context"
	"github.com/ory/kratos/ui/node"
)

var ctx = context.Background()

func sortNodes(n node.Nodes) error {
	return n.SortBySchema(ctx,
		node.SortByGroups([]node.Group{
			node.DefaultGroup,
			node.OpenIDConnectGroup,
			node.PasswordGroup,
			node.WebAuthnGroup,
			node.TOTPGroup,
			node.LookupGroup,
		}),
		node.SortUseOrder([]string{
			"password_identifier",
			"password",
		}),
	)
}
