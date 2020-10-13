package x

import (
	"github.com/markbates/pkger"

	"github.com/ory/x/logrusx"
	"github.com/ory/x/viperx"
)

func WatchAndValidateViper(log *logrusx.Logger) {
	schema := MustPkgerRead(pkger.Open("/.schema/config.schema.json"))
	viperx.WatchAndValidateViper(log, schema,
		"ORY Kratos", []string{"serve", "profiling", "log"}, "")
}
