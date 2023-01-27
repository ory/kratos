// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"github.com/ory/kratos/text"
	"github.com/ory/x/stringsx"

	"github.com/ory/kratos/ui/container"

	"github.com/ory/kratos/ui/node"

	"github.com/gofrs/uuid"
)

type FlowMethod struct {
	*container.Container
}

func AddProviders(c *container.Container, providers []Configuration, message func(provider string) *text.Message) {
	for _, p := range providers {
		AddProvider(c, p.ID, message(
			stringsx.Coalesce(p.Label, p.ID)))
	}
}

func AddProvider(c *container.Container, providerID string, message *text.Message) {
	c.GetNodes().Append(
		node.NewInputField("provider", providerID, node.OpenIDConnectGroup, node.InputAttributeTypeSubmit).WithMetaLabel(message),
	)
}

func NewFlowMethod(f *container.Container) *FlowMethod {
	return &FlowMethod{Container: f}
}

type ider interface {
	GetID() uuid.UUID
}
