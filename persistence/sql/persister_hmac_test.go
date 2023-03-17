// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"os"
	"testing"

	"github.com/ory/x/contextx"

	"github.com/ory/x/configx"
	"github.com/ory/x/otelx"

	"github.com/gobuffalo/pop/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
)

type logRegistryOnly struct {
	l *logrusx.Logger
	c *config.Config
}

func (l *logRegistryOnly) Config() *config.Config {
	return l.c
}

func (l *logRegistryOnly) Contextualizer() contextx.Contextualizer {
	//TODO implement me
	panic("implement me")
}

func (l *logRegistryOnly) Logger() *logrusx.Logger {
	if l.l == nil {
		l.l = logrusx.New("kratos", "testing")
	}
	return l.l
}

func (l *logRegistryOnly) Audit() *logrusx.Logger {
	panic("implement me")
}

func (l *logRegistryOnly) Tracer(ctx context.Context) *otelx.Tracer {
	return otelx.NewNoop(l.l, new(otelx.Config))
}
func (l *logRegistryOnly) IdentityTraitsSchemas(ctx context.Context) (schema.Schemas, error) {
	panic("implement me")
}

func (l *logRegistryOnly) IdentityValidator() *identity.Validator {
	panic("implement me")
}

var _ persisterDependencies = &logRegistryOnly{}

func TestPersisterHMAC(t *testing.T) {
	ctx := context.Background()
	conf := config.MustNew(t, logrusx.New("", ""), os.Stderr, configx.SkipValidation())
	conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"foobarbaz"})
	c, err := pop.NewConnection(&pop.ConnectionDetails{URL: "sqlite://foo?mode=memory"})
	require.NoError(t, err)
	p, err := NewPersister(context.Background(), &logRegistryOnly{c: conf}, c)
	require.NoError(t, err)

	assert.True(t, p.hmacConstantCompare(context.Background(), "hashme", p.hmacValue(context.Background(), "hashme")))
	assert.False(t, p.hmacConstantCompare(context.Background(), "notme", p.hmacValue(context.Background(), "hashme")))
	assert.False(t, p.hmacConstantCompare(context.Background(), "hashme", p.hmacValue(context.Background(), "notme")))

	hash := p.hmacValue(context.Background(), "hashme")
	conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"notfoobarbaz"})
	assert.False(t, p.hmacConstantCompare(context.Background(), "hashme", hash))
	assert.True(t, p.hmacConstantCompare(context.Background(), "hashme", p.hmacValue(context.Background(), "hashme")))

	conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"notfoobarbaz", "foobarbaz"})
	assert.True(t, p.hmacConstantCompare(context.Background(), "hashme", hash))
	assert.True(t, p.hmacConstantCompare(context.Background(), "hashme", p.hmacValue(context.Background(), "hashme")))
	assert.NotEqual(t, hash, p.hmacValue(context.Background(), "hashme"))
}
