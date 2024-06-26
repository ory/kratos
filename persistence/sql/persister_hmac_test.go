// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"os"
	"testing"

	confighelpers "github.com/ory/kratos/driver/config/testhelpers"

	"github.com/ory/x/configx"

	"github.com/ory/x/contextx"

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

func (l *logRegistryOnly) Tracer(context.Context) *otelx.Tracer {
	return otelx.NewNoop(l.l, new(otelx.Config))
}
func (l *logRegistryOnly) IdentityTraitsSchemas(context.Context) (schema.IdentitySchemaList, error) {
	panic("implement me")
}

func (l *logRegistryOnly) IdentityValidator() *identity.Validator {
	panic("implement me")
}

var _ persisterDependencies = &logRegistryOnly{}

func TestPersisterHMAC(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	baseSecret := "foobarbaz"
	baseSecretBytes := []byte(baseSecret)
	opts := []configx.OptionModifier{configx.SkipValidation(), configx.WithValue(config.ViperKeySecretsDefault, []string{baseSecret})}
	conf := config.MustNew(t, logrusx.New("", ""), os.Stderr, &confighelpers.TestConfigProvider{Contextualizer: &contextx.Default{}, Options: opts}, opts...)
	c, err := pop.NewConnection(&pop.ConnectionDetails{URL: "sqlite://foo?mode=memory"})
	require.NoError(t, err)
	p, err := NewPersister(ctx, &logRegistryOnly{c: conf}, c)
	require.NoError(t, err)

	t.Run("case=behaves deterministically", func(t *testing.T) {
		assert.Equal(t, hmacValueWithSecret(ctx, "hashme", baseSecretBytes), p.hmacValue(ctx, "hashme"))
		assert.NotEqual(t, hmacValueWithSecret(ctx, "notme", baseSecretBytes), p.hmacValue(ctx, "hashme"))
		assert.NotEqual(t, hmacValueWithSecret(ctx, "hashme", baseSecretBytes), p.hmacValue(ctx, "notme"))
	})

	hash := p.hmacValue(ctx, "hashme")
	newSecret := "not" + baseSecret

	t.Run("case=with only new sectet", func(t *testing.T) {
		ctx = confighelpers.WithConfigValue(ctx, config.ViperKeySecretsDefault, []string{newSecret})
		assert.NotEqual(t, hmacValueWithSecret(ctx, "hashme", baseSecretBytes), p.hmacValue(ctx, "hashme"))
		assert.Equal(t, hmacValueWithSecret(ctx, "hashme", []byte(newSecret)), p.hmacValue(ctx, "hashme"))
	})

	t.Run("case=with new and old secret", func(t *testing.T) {
		ctx = confighelpers.WithConfigValue(ctx, config.ViperKeySecretsDefault, []string{newSecret, baseSecret})
		assert.Equal(t, hmacValueWithSecret(ctx, "hashme", []byte(newSecret)), p.hmacValue(ctx, "hashme"))
		assert.NotEqual(t, hash, p.hmacValue(ctx, "hashme"))
	})
}
