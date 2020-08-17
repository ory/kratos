package x

import (
	"github.com/gobuffalo/packr/v2"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/viperx"
)

var schemas = packr.New("schemas", "../.schema")

func WatchAndValidateViper(log *logrusx.Logger) {
	schema, err := schemas.Find("config.schema.json")
	if err != nil {
		log.WithError(err).Fatal("Unable to open configuration JSON Schema.")
	}

	viperx.WatchAndValidateViper(log, schema, "ORY Kratos", []string{"serve", "profiling", "log"}, "")
}
