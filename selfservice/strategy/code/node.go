// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
)

var nodeSubmitRegistration = node.NewInputField("method", identity.CredentialsTypeCodeAuth, node.CodeGroup, node.InputAttributeTypeSubmit).
	WithMetaLabel(text.NewInfoSelfServiceRegistrationRegisterCode())

var nodeRegistrationResendNode = node.NewInputField("resend", identity.CredentialsTypeCodeAuth, node.CodeGroup, node.InputAttributeTypeSubmit).
	WithMetaLabel(text.NewInfoNodeResendOTP())

var nodeRegistrationSelectCredentialsNode = node.NewInputField(
	"screen",
	"credential-selection",
	node.ProfileGroup,
	node.InputAttributeTypeSubmit,
).WithMetaLabel(text.NewInfoRegistrationBack())

var nodeContinueButton = node.NewInputField("method", identity.CredentialsTypeCodeAuth, node.CodeGroup, node.InputAttributeTypeSubmit).
	WithMetaLabel(text.NewInfoNodeLabelContinue())

var nodeCodeInputFieldHidden = node.NewInputField("method", identity.CredentialsTypeCodeAuth, node.CodeGroup, node.InputAttributeTypeHidden)

func nodeCodeInputField() *node.Node {
	return node.NewInputField("code", nil, node.CodeGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute)
}
