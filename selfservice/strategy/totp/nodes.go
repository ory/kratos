package totp

import (
	"github.com/pquerna/otp"

	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
)

func NewVerifyTOTPNode() *node.Node {
	return node.NewInputField("verification_totp", nil, node.TOTPGroup,
		node.InputAttributeTypeText,
		node.WithRequiredInputAttribute).
		WithMetaLabel(text.NewInfoNodeLabelVerifyOTP())
}

func NewTOTPImageQRNode(key *otp.Key) (*node.Node, error) {
	src, err := KeyToHTMLImage(key)
	if err != nil {
		return nil, err
	}

	return node.NewImageField("totp_key_qr", src, node.TOTPGroup).
		WithMetaLabel(text.NewInfoSelfServiceSettingsTOTPQRCode()), nil
}

func NewTOTPSourceURLNode(key *otp.Key) *node.Node {
	return node.NewTextField("totp_key_secret", text.NewInfoSelfServiceSettingsTOTPSecret(key.Secret()), node.TOTPGroup)
}

func NewUnlinkTOTPNode() *node.Node {
	return node.NewInputField("unlink_totp", "true", node.TOTPGroup,
		node.InputAttributeTypeSubmit,
		node.WithRequiredInputAttribute).
		WithMetaLabel(text.NewInfoSelfServiceSettingsUpdateUnlinkTOTP())
}
