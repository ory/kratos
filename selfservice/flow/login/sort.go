package login

import "github.com/ory/kratos/ui/node"

func sortNodes(n node.Nodes) error {
	return n.SortBySchema(
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
