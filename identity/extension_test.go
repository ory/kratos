package identity_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/x"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	. "github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
)

func TestValidationExtension(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)
	routera := x.NewRouterAdmin()
	routerp := x.NewRouterPublic()
	reg.SchemaHandler().RegisterAdminRoutes(routera)
	reg.SchemaHandler().RegisterPublicRoutes(routerp)

	tsa := httptest.NewServer(routera)
	tsp := httptest.NewServer(routerp)
	defer tsa.Close()
	defer tsp.Close()

	conf := internal.NewConfigurationWithDefaults()
	viper.Set(configuration.ViperKeyURLsSelfAdmin, tsa.URL)
	viper.Set(configuration.ViperKeyURLsSelfPublic, tsp.URL)
	_, _ = reg.SchemaPersister().RegisterDefaultSchema("file://../identity/stub/extension.schema.json")

	v := NewValidator(conf, reg)

	i := NewIdentity(uuid.Nil)
	i.Traits = Traits(`{
  "email": "foo@bar.com",
  "names": [
    "foobar",
    "bazbar"
  ],
  "age": 1
}`)
	require.NoError(t, v.Validate(i))

	c, ok := i.GetCredentials(CredentialsTypePassword)
	require.True(t, ok)
	assert.ElementsMatch(t, []string{"foo@bar.com", "foobar", "bazbar"}, c.Identifiers)
}
