package errorx

import "github.com/ory/kratos/x"

type Registry interface {
	ErrorManager() Manager
	x.WriterProvider
}
