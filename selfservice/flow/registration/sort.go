package registration

import "github.com/ory/kratos/ui/node"

func SortNodes(n node.Nodes, schemaRef string) error {
	return n.SortBySchema(
		node.SortBySchema(schemaRef),
		node.SortByGroups([]node.Group{
			node.DefaultGroup,
			node.OpenIDConnectGroup,
			node.PasswordGroup,
		}),
		node.SortUpdateOrder(node.PasswordLoginOrder),
	)
}
