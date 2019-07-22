package errorx

import "github.com/ory/hive-cloud/hive/x"

type Registry interface {
	ErrorManager() Manager
	x.WriterProvider
}
