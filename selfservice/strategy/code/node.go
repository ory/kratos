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
