// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
)

func NewLinkNode(providerID, providerLabel string) *node.Node {
	return node.NewInputField("link", providerID, node.OpenIDConnectGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoSelfServiceSettingsUpdateLinkOIDC(providerLabel))
}

func NewUnlinkNode(providerID, providerLabel string) *node.Node {
	return node.NewInputField("unlink", providerID, node.OpenIDConnectGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoSelfServiceSettingsUpdateUnlinkOIDC(providerLabel))
}
