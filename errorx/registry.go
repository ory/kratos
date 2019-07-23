package errorx

import "github.com/ory/hive/x"

type Registry interface {
	ErrorManager() Manager
	x.WriterProvider
}
