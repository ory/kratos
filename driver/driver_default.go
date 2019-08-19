package driver

import (
	"github.com/ory/x/logrusx"
	"github.com/sirupsen/logrus"

	"github.com/ory/hive/driver/configuration"
)

type DefaultDriver struct {
	l logrus.FieldLogger
	c configuration.Provider
	r Registry
}

func NewDefaultDriver(l logrus.FieldLogger, version, build, date string) Driver {
	if l == nil {
		l = logrusx.New()
	}

	c := configuration.NewViperProvider(l)

	r, err := NewRegistry(c)
	if err != nil {
		l.WithError(err).Fatal("Unable to instantiate service registry.")
	}

	r.
		WithConfig(c).
		WithLogger(l).
		WithBuildInfo(version, build, date)

	// Init forces the driver to initialize and circumvent lazy loading issues.
	if err = r.Init(); err != nil {
		l.WithError(err).Fatal("Unable to initialize service registry.")
	}

	return &DefaultDriver{r: r, c: c}
}

func (r *DefaultDriver) BuildInfo() *BuildInfo {
	return &BuildInfo{}
}

func (r *DefaultDriver) Logger() logrus.FieldLogger {
	if r.l == nil {
		r.l = logrusx.New()
	}
	return r.l
}

func (r *DefaultDriver) Configuration() configuration.Provider {
	return r.c
}

func (r *DefaultDriver) Registry() Registry {
	return r.r
}
