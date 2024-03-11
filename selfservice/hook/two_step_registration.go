// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

var _ registration.PreHookExecutor = new(TwoStepRegistration)

type (
	twoStepRegistrationDeps interface {
		x.WriterProvider
		config.Provider
	}

	TwoStepRegistration struct {
		d twoStepRegistrationDeps
	}
)

func NewTwoStepRegistration(d twoStepRegistrationDeps) *TwoStepRegistration {
	return &TwoStepRegistration{d: d}
}

func (e *TwoStepRegistration) ExecuteRegistrationPreHook(_ http.ResponseWriter, _ *http.Request, regFlow *registration.Flow) (err error) {
	stepOneNodes := make([]*node.Node, 0, len(regFlow.UI.Nodes))
	stepTwoNodes := make([]*node.Node, 0, len(regFlow.UI.Nodes))
	for _, n := range regFlow.UI.Nodes {
		if n.Group == node.ProfileGroup || n.Group == node.OpenIDConnectGroup || n.Group == node.DefaultGroup {
			stepOneNodes = append(stepOneNodes, n)
		} else {
			stepTwoNodes = append(stepTwoNodes, n)
		}
	}

	regFlow.UI.Nodes = stepOneNodes

	regFlow.InternalContext, err = sjson.SetBytes(regFlow.InternalContext, "stepTwoNodes", stepTwoNodes)
	if err != nil {
		return errors.WithStack(err)
	}
	regFlow.InternalContext, err = sjson.SetBytes(regFlow.InternalContext, "stepOneNodes", stepOneNodes)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
