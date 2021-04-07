package settings

import "github.com/ory/kratos/ui/node"

func sortNodes(n node.Nodes, schemaRef string) error {
	return n.SortBySchema(
		node.SortBySchema(schemaRef),
		node.SortByGroups([]node.Group{
			node.DefaultGroup,
			node.ProfileGroup,
			node.PasswordGroup,
			node.OpenIDConnectGroup,
		}),
	)
}
