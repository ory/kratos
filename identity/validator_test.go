package identity_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	. "github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
)

func TestSchemaValidator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	router := httprouter.New()
	router.GET("/schema/:name", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		_, _ = w.Write([]byte(`{
  "$id": "https://example.com/person.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
    "` + ps.ByName("name") + `": {
      "type": "string",
      "description": "The person's first name."
    },
    "lastName": {
      "type": "string",
      "description": "The person's last name."
    },
    "age": {
      "description": "Age in years which must be equal to or greater than zero.",
      "type": "integer",
      "minimum": 1
    }
  },
  "additionalProperties": false
}`))
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	conf, reg := internal.NewFastRegistryWithMocks(t)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, ts.URL+"/schema/firstName")
	viper.Set(configuration.ViperKeyIdentityTraitsSchemas, []configuration.SchemaConfig{
		{ID: "whatever", URL: ts.URL + "/schema/whatever"},
		{ID: "unreachable-url", URL: ts.URL + "/404-not-found"},
	})
	v := NewValidator(reg, conf)

	for k, tc := range []struct {
		i   *Identity
		err string
	}{
		{
			i: &Identity{
				Traits: Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			},
		},
		{
			i: &Identity{
				Traits: Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": -1 }`),
			},
			err: "I[#/traits/age] S[#/properties/age/minimum] must be >= 1 but found -1",
		},
		{
			i: &Identity{
				Traits: Traits(`{ "whatever": "first-name", "lastName": "last-name", "age": 1 }`),
			},
			err: `I[#/traits] S[#/additionalProperties] additionalProperties "whatever" not allowed`,
		},
		{
			i: &Identity{
				TraitsSchemaID: "whatever",
				Traits:         Traits(`{ "whatever": "first-name", "lastName": "last-name", "age": 1 }`),
			},
		},
		{
			i: &Identity{
				TraitsSchemaID: "whatever",
				Traits:         Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			},
			err: `I[#/traits] S[#/additionalProperties] additionalProperties "firstName" not allowed`,
		},
		{
			i: &Identity{
				TraitsSchemaID: "unreachable-url",
				Traits:         Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			},
			err: "An internal server error occurred, please contact the system administrator",
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			err := v.Validate(tc.i)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.err)
			}
		})
	}
}
