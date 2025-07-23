// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"io"

	"github.com/ory/x/servicelocatorx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/logrusx"
)

func New(ctx context.Context, stdOutOrErr io.Writer, dOpts ...RegistryOption) (*RegistryDefault, error) {
	r, err := NewWithoutInit(ctx, stdOutOrErr, dOpts...)
	if err != nil {
		return nil, err
	}

	ctxter := r.Contextualizer()
	if err := r.Init(ctx, ctxter, dOpts...); err != nil {
		r.Logger().WithError(err).Error("Unable to initialize service registry.")
		return nil, err
	}

	return r, nil
}

func NewWithoutInit(ctx context.Context, stdOutOrErr io.Writer, dOpts ...RegistryOption) (*RegistryDefault, error) {
	opts := newOptions(dOpts)
	sl := servicelocatorx.NewOptions(opts.serviceLocatorOptions...)

	l := sl.Logger()
	if l == nil {
		l = logrusx.New("Ory Kratos", config.Version)
	}

	c := opts.config
	if c == nil {
		var err error
		c, err = config.New(ctx, l, stdOutOrErr, sl.Contextualizer(), opts.configOptions...)
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
	r.slOptions = sl
	r.SetContextualizer(sl.Contextualizer())

	return r, nil
}
