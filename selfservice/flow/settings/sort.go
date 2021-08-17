package settings

import (
	"github.com/ory/kratos/ui/node"
)

func sortNodes(n node.Nodes, schemaRef string) error {
	return n.SortBySchema(
		node.SortBySchema(schemaRef),
		node.SortByGroups([]node.Group{
			node.DefaultGroup,
			node.ProfileGroup,
			node.PasswordGroup,
			node.OpenIDConnectGroup,
			node.LookupGroup,
			node.WebAuthnGroup,
			node.TOTPGroup,
		}),
		node.SortUseOrder([]string{
			// Lookup
			node.LookupReveal,
			node.LookupRegenerate,
			node.LookupCodes,
			node.LookupConfirm,

			// Lookup
			node.WebAuthnRemove,
			node.WebAuthnRegisterDisplayName,
			node.WebAuthnRegister,

			// TOTP
			node.TOTPSecretKey,
			node.TOTPQR,
			node.TOTPUnlink,
			node.TOTPCode,
		}),
	)
}
