package x

import (
	"io/ioutil"

	"github.com/markbates/pkger"

	"github.com/ory/x/logrusx"
	"github.com/ory/x/viperx"
)

func WatchAndValidateViper(log *logrusx.Logger) {
	file, err := pkger.Open("/.schema/config.schema.json")
	if err != nil {
		log.WithError(err).Fatal("Unable to open configuration JSON Schema.")
	}
	defer file.Close()
	schema, err := ioutil.ReadAll(file)
	if err != nil {
		log.WithError(err).Fatal("Unable to read configuration JSON Schema.")
	}
	viperx.WatchAndValidateViper(log, schema, "ORY Kratos", []string{"serve", "profiling", "log"})
}
