package driver

import (
	"context"
	"io"

	"github.com/ory/kratos/x/servicelocatorx"
	"github.com/ory/x/contextx"
	"github.com/ory/x/servicelocator"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
)

func New(ctx context.Context, stdOutOrErr io.Writer, opts ...configx.OptionModifier) (Registry, error) {
	r, err := NewWithoutInit(ctx, stdOutOrErr, opts...)
	if err != nil {
		return nil, err
	}

	ctxter := servicelocator.Contextualizer(ctx, &contextx.Default{})
	if err := r.Init(ctx, ctxter); err != nil {
		r.Logger().WithError(err).Error("Unable to initialize service registry.")
		return nil, err
	}

	return r, nil
}

func NewWithoutInit(ctx context.Context, stdOutOrErr io.Writer, opts ...configx.OptionModifier) (Registry, error) {
	l := logrusx.New("Ory Kratos", config.Version)

	c := servicelocatorx.ConfigFromContext(ctx, nil)
	if c == nil {
		var err error
		c, err = config.New(ctx, l, stdOutOrErr, opts...)
		if err != nil {
			l.WithError(err).Error("Unable to instantiate configuration.")
			return nil, err
		}
	}

	r, err := NewRegistryFromDSN(ctx, c, l)
	if err != nil {
		l.WithError(err).Error("Unable to instantiate service registry.")
		return nil, err
	}

	c.SetTracer(ctx, r.Tracer(ctx))
	return r, nil
}
