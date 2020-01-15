package driver

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/driver/configuration"
)

type DefaultDriver struct {
	c configuration.Provider
	r Registry
}

func NewDefaultDriver(l logrus.FieldLogger, version, build, date string, dev bool) (Driver, error) {
	if l == nil {
		l = logrusx.New()
	}

	c := configuration.NewViperProvider(l, dev)

	r, err := NewRegistry(c)
	if err != nil {
		return nil, errors.Wrap(err, "unable to instantiate service registry")
	}

	r.
		WithConfig(c).
		WithLogger(l).
		WithBuildInfo(version, build, date)

	// Init forces the driver to initialize and circumvent lazy loading issues.
	if err = r.Init(); err != nil {
		return nil, errors.Wrap(err, "unable to initialize service registry")
	}

	return &DefaultDriver{r: r, c: c}, nil
}

func MustNewDefaultDriver(l logrus.FieldLogger, version, build, date string, dev bool) Driver {
	d, err := NewDefaultDriver(l, version, build, date, dev)
	if err != nil {
		l.WithError(err).Fatal("Unable to initialize driver.")
	}
	return d
}

func (r *DefaultDriver) BuildInfo() *BuildInfo {
	return &BuildInfo{}
}

func (r *DefaultDriver) Logger() logrus.FieldLogger {
	if r.r == nil {
		return logrusx.New()
	}
	return r.r.Logger()
}

func (r *DefaultDriver) Configuration() configuration.Provider {
	return r.c
}

func (r *DefaultDriver) Registry() Registry {
	return r.r
}
