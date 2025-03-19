// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/text"
	"github.com/ory/x/stringsx"

	"github.com/ory/kratos/ui/container"

	"github.com/ory/kratos/ui/node"

	"github.com/gofrs/uuid"
)

type FlowMethod struct {
	*container.Container
}

func AddProviders(c *container.Container, providers []Configuration, message func(provider string, providerId string) *text.Message, credentialsType identity.CredentialsType) {
	for _, p := range providers {
		if len(p.OrganizationID) > 0 {
			continue
		}
		AddProvider(c, p.ID, message(stringsx.Coalesce(p.Label, p.ID), p.ID), credentialsType)
	}
}

func AddProvider(c *container.Container, providerID string, message *text.Message, credentialsType identity.CredentialsType) {
	group := node.OpenIDConnectGroup
	if credentialsType == identity.CredentialsTypeSAML {
		group = node.SAMLGroup
	}
	field := node.NewInputField("provider", providerID, group, node.InputAttributeTypeSubmit).WithMetaLabel(message)
	c.GetNodes().RemoveMatching(field)
	c.GetNodes().Append(field)
}

func NewFlowMethod(f *container.Container) *FlowMethod {
	return &FlowMethod{Container: f}
}

type ider interface {
	GetID() uuid.UUID
}
