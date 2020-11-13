package x

import (
	"github.com/markbates/pkger"

	"github.com/ory/x/pkgerx"

	"github.com/ory/x/logrusx"
	"github.com/ory/x/viperx"
)

func WatchAndValidateViper(log *logrusx.Logger) {
	schema := pkgerx.MustRead(pkger.Open("/.schema/config.schema.json"))
	viperx.WatchAndValidateViper(log, schema,
		"ORY Kratos", []string{"serve", "profiling", "log"}, "")
}
