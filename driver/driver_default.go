package driver

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/driver/configuration"
)

type DefaultDriver struct {
	c configuration.Provider
	r Registry
}

// IsSQLiteMemoryMode returns true if SQLite if configured to use memory mode
func IsSQLiteMemoryMode(dsn string) bool {
	/*
		if urlParts := strings.SplitN(dsn, "?", 2); len(urlParts) == 2 && strings.HasPrefix(dsn, "sqlite") {
			queryVals, err := url.ParseQuery(urlParts[1])
			fmt.Println("2>>>>>> " + urlParts[1])
			if err == nil && queryVals.Get("mode") == "memory" {
				return true
			}
		}
		return false
	*/
	if dsn == configuration.DefaultSQLiteMemoryDSN {
		return true
	}
	return false
}

func NewDefaultDriver(l *logrusx.Logger, version, build, date string, dev bool) (Driver, error) {
	if l == nil {
		l = logrusx.New("ORY Kratos", version)
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

	dsn := c.DSN()
	// if dsn is memory we have to run the migrations on every start
	if IsSQLiteMemoryMode(dsn) {
		l.Print("Kratos is running migrations on every startup as DSN is memory.\n")
		l.Print("This means your data is lost when Kratos terminates.\n")
		if err := r.Persister().MigrateUp(context.Background()); err != nil {
			return nil, err
		}
	}
	return &DefaultDriver{r: r, c: c}, nil
}

func MustNewDefaultDriver(l *logrusx.Logger, version, build, date string, dev bool) Driver {
	d, err := NewDefaultDriver(l, version, build, date, dev)
	if err != nil {
		l.WithError(err).Fatal("Unable to initialize driver.")
	}
	return d
}

func (r *DefaultDriver) BuildInfo() *BuildInfo {
	return &BuildInfo{}
}

func (r *DefaultDriver) Logger() *logrusx.Logger {
	if r.r == nil {
		return logrusx.New("ORY Kratos", r.BuildInfo().Version)
	}
	return r.r.Logger()
}

func (r *DefaultDriver) Configuration() configuration.Provider {
	return r.c
}

func (r *DefaultDriver) Registry() Registry {
	return r.r
}
