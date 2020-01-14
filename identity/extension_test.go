package identity_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ory/kratos/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	. "github.com/ory/kratos/identity"
)

func TestValidationExtension(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("stub")))
	defer ts.Close()

	_, reg := internal.NewRegistryDefault(t)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, ts.URL+"/extension.schema.json")
	v := NewValidator(reg)

	i := NewIdentity("")
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
