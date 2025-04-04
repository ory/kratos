// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package webauthn

import (
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x/webauthnx"
)

func nodeWebauthnRegistrationOptions(opts []byte) *node.Node {
	return webauthnx.NewWebAuthnConnectionTrigger(string(opts)).WithMetaLabel(text.NewInfoSelfServiceRegistrationRegisterWebAuthn())
}

var nodeDisplayName = webauthnx.NewWebAuthnConnectionName()
var nodeConnectionInput = webauthnx.NewWebAuthnConnectionInput()
