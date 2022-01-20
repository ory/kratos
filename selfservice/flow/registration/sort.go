package registration

import (
	"context"

	"github.com/ory/kratos/ui/node"
)

var ctx = context.Background()

func SortNodes(n node.Nodes, schemaRef string) error {
	return n.SortBySchema(ctx,
		node.SortBySchema(schemaRef),
		node.SortByGroups([]node.Group{
			node.DefaultGroup,
			node.OpenIDConnectGroup,
			node.PasswordGroup,
		}),
		node.SortUpdateOrder(node.PasswordLoginOrder),
	)
}
