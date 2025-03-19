// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
)

var (
	nodePreviousScreen = node.NewInputField(
		"screen",
		"previous",
		node.ProfileGroup,
		node.InputAttributeTypeSubmit,
	).WithMetaLabel(text.NewInfoRegistrationBack())
	nodeSubmitProfile = node.NewInputField(
		"method",
		"profile",
		node.ProfileGroup,
		node.InputAttributeTypeSubmit,
	).WithMetaLabel(text.NewInfoRegistration())
)
