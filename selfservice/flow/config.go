package flow

import "github.com/ory/kratos/selfservice/form"

// swagger:ignore
type MethodConfigurator interface {
	form.NodeGetter

	form.ErrorParser

	// form.NodeSetter
	// form.NodeUnsetter
	form.ValueSetter

	form.Resetter
	form.MessageResetter
	form.CSRFSetter
	form.FieldSorter
}
