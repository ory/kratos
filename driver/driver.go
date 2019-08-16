package driver

import (
	"github.com/sirupsen/logrus"

	"github.com/ory/hive/driver/configuration"
)

type BuildInfo struct {
	Version string
	Hash    string
	Time    string
}

type Driver interface {
	Logger() logrus.FieldLogger
	Configuration() configuration.Provider
	Registry() Registry
}
