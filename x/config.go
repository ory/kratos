package x

import (
	"github.com/markbates/pkger"

	"github.com/ory/x/logrusx"
	"github.com/ory/x/viperx"
)

var schema = MustPkgerRead(pkger.Open("/.schema/config.schema.json"))

func WatchAndValidateViper(log *logrusx.Logger) {
	viperx.WatchAndValidateViper(log, schema, "ORY Kratos", []string{"serve", "profiling", "log"}, "")
}
