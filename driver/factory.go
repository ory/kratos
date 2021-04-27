package driver

import (
	"context"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
)

func New(ctx context.Context, opts ...configx.OptionModifier) Registry {
	l := logrusx.New("Ory Kratos", config.Version)
	c, err := config.New(ctx, l, opts...)
	if err != nil {
		l.WithError(err).Fatal("Unable to instantiate configuration.")
	}

	r, err := NewRegistryFromDSN(c, l)
	if err != nil {
		l.WithError(err).Fatal("Unable to instantiate service registry.")
	}

	if err = r.Init(ctx); err != nil {
		l.WithError(err).Fatal("Unable to initialize service registry.")
	}

	c.Source().SetTracer(ctx, r.Tracer(ctx))

	return r
}
