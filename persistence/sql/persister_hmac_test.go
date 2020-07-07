package sql

import (
	"testing"

	"github.com/gobuffalo/pop/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"
	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/driver/configuration"
)

func TestPersisterHMAC(t *testing.T) {
	viper.Set(configuration.ViperKeySecretsDefault, []string{"foobarbaz"})
	c, err := pop.NewConnection(&pop.ConnectionDetails{URL: "sqlite://foo?mode=memory"})
	require.NoError(t, err)
	p, err := NewPersister(nil, configuration.NewViperProvider(logrusx.New("", ""), false), c)
	require.NoError(t, err)

	assert.True(t, p.hmacConstantCompare("hashme", p.hmacValue("hashme")))
	assert.False(t, p.hmacConstantCompare("notme", p.hmacValue("hashme")))
	assert.False(t, p.hmacConstantCompare("hashme", p.hmacValue("notme")))

	hash := p.hmacValue("hashme")
	viper.Set(configuration.ViperKeySecretsDefault, []string{"notfoobarbaz"})
	assert.False(t, p.hmacConstantCompare("hashme", hash))
	assert.True(t, p.hmacConstantCompare("hashme", p.hmacValue("hashme")))

	viper.Set(configuration.ViperKeySecretsDefault, []string{"notfoobarbaz", "foobarbaz"})
	assert.True(t, p.hmacConstantCompare("hashme", hash))
	assert.True(t, p.hmacConstantCompare("hashme", p.hmacValue("hashme")))
	assert.NotEqual(t, hash, p.hmacValue("hashme"))
}
