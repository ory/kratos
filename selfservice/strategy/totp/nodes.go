package totp

import (
	"github.com/pquerna/otp"

	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
)

func NewVerifyTOTPNode() *node.Node {
	return node.NewInputField(node.TOTPCode, nil, node.TOTPGroup,
		node.InputAttributeTypeText,
		node.WithRequiredInputAttribute).
		WithMetaLabel(text.NewInfoNodeLabelVerifyOTP())
}

func NewTOTPImageQRNode(key *otp.Key) (*node.Node, error) {
	src, err := KeyToHTMLImage(key)
	if err != nil {
		return nil, err
	}

	return node.NewImageField(node.TOTPQR, src, node.TOTPGroup).
		WithMetaLabel(text.NewInfoSelfServiceSettingsTOTPQRCode()), nil
}

func NewTOTPSourceURLNode(key *otp.Key) *node.Node {
	return node.NewTextField(node.TOTPSecretKey,
		text.NewInfoSelfServiceSettingsTOTPSecret(key.Secret()), node.TOTPGroup).
		WithMetaLabel(text.NewInfoSelfServiceSettingsTOTPSecretLabel())
}

func NewUnlinkTOTPNode() *node.Node {
	return node.NewInputField(node.TOTPUnlink, "true", node.TOTPGroup,
		node.InputAttributeTypeSubmit,
		node.WithRequiredInputAttribute).
		WithMetaLabel(text.NewInfoSelfServiceSettingsUpdateUnlinkTOTP())
}
