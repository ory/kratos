package registration

import (
	"context"

	"github.com/ory/kratos/ui/node"
)

func SortNodes(ctx context.Context, n node.Nodes, schemaRef string) error {
	return n.SortBySchema(ctx,
		node.SortBySchema(schemaRef),
		node.SortByGroups([]node.UiNodeGroup{
			node.DefaultGroup,
			node.OpenIDConnectGroup,
			node.WebAuthnGroup,
			node.PasswordGroup,
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
