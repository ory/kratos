package identity_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/hive/driver/configuration"
	. "github.com/ory/hive/identity"
	"github.com/ory/hive/internal"
)

func TestValidationExtension(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("stub")))
	defer ts.Close()

	conf := internal.NewConfigurationWithDefaults()
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, ts.URL+"/extension.schema.json")
	v := NewValidator(conf)

	i := NewIdentity("")
	i.Traits = json.RawMessage(`{
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
