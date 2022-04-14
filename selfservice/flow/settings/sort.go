package settings

import (
	"context"

	"github.com/ory/kratos/ui/node"
)

func sortNodes(ctx context.Context, n node.Nodes, schemaRef string) error {
	return n.SortBySchema(ctx,
		node.SortBySchema(schemaRef),
		node.SortByGroups([]node.UiNodeGroup{
			node.DefaultGroup,
			node.ProfileGroup,
			node.PasswordGroup,
			node.OpenIDConnectGroup,
			node.LookupGroup,
			node.WebAuthnGroup,
			node.TOTPGroup,
		}),
		node.SortUseOrderAppend([]string{
			// Lookup
			node.LookupReveal,
			node.LookupRegenerate,
			node.LookupCodes,
			node.LookupConfirm,

			// WebAuthn
			node.WebAuthnRemove,
			node.WebAuthnRegisterDisplayName,
			node.WebAuthnRegister,

			// TOTP
			node.TOTPQR,
			node.TOTPSecretKey,
			node.TOTPUnlink,
			node.TOTPCode,
		}),
	)
}
