// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import (
	"github.com/ory/kratos/ui/container"
)

// swagger:ignore
type MethodConfigurator interface {
	container.NodeGetter

	container.ErrorParser

	// form.NodeSetter
	// form.NodeUnsetter
	container.ValueSetter

	container.Resetter
	container.MessageResetter
	container.CSRFSetter
	container.FieldSorter
}
