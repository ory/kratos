package saml

import (
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/stringsx"
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
		node.NewInputField("samlProvider", providerID, node.SAMLGroup, node.InputAttributeTypeSubmit).WithMetaLabel(message),
	)
}

func NewFlowMethod(f *container.Container) *FlowMethod {
	return &FlowMethod{Container: f}
}
