package schema

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"

	"github.com/ory/gojsonschema"
	"github.com/ory/x/stringsx"
)

func TestSchemaValidator(t *testing.T) {
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

	for k, tc := range []struct {
		i   json.RawMessage
		err string
		u   string
	}{
		{
			i: json.RawMessage(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
		},
		{
			i:   json.RawMessage(`{ "firstName": "first-name", "lastName": "last-name", "age": -1 }`),
			err: "must be greater than or equal to 1",
			// gojsonschema.ResultError{
			// 	Field: "age", Type: "number_gte", Value: "-1", Message: "Must be greater than or equal to 1/1",
			// 	Details: map[string]interface{}{"min": new(big.Rat).SetInt(big.NewInt(1)), "field": "age", "context": "(root).age"},
			// },
		},
		{
			i:   json.RawMessage(`{ "whatever": "first-name", "lastName": "last-name", "age": 1 }`),
			err: "additional property whatever is not allowed",
			// gojsonschema.ResultError{
			// 	Field:   "(root)",
			// 	Type:    "additional_property_not_allowed",
			// 	Message: "Additional property whatever is not allowed",
			// 	Value:   "first-name",
			// 	Details: map[string]interface{}{"property": "whatever", "field": "(root)", "context": "(root)"},
			// },
		},
		{
			u: ts.URL + "/schema/whatever",
			i: json.RawMessage(`{ "whatever": "first-name", "lastName": "last-name", "age": 1 }`),
		},
		{
			u:   ts.URL + "/schema/whatever",
			i:   json.RawMessage(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			err: "additional property firstName is not allowed",
			// gojsonschema.ResultError{
			// 	Field:   "(root)",
			// 	Type:    "additional_property_not_allowed",
			// 	Message: "Additional property firstName is not allowed",
			// 	Value:   "first-name",
			// 	Details: map[string]interface{}{"property": "firstName", "field": "(root)", "context": "(root)"},
			// },
		},
		{
			u:   ts.URL,
			i:   json.RawMessage(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			err: "An internal server error occurred, please contact the system administrator",
		},
		{
			u:   "not-a-url",
			i:   json.RawMessage(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			err: "An internal server error occurred, please contact the system administrator",
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			err := NewValidator().Validate(
				stringsx.Coalesce(
					tc.u,
					ts.URL+"/schema/firstName"),
				gojsonschema.NewGoLoader(tc.i),
			)

			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.err)
			}

			// require.Len(t, fe, len(tc.expected))
			// for _, g := range fe {
			// 	var found bool
			// 	for _, e := range tc.expected {
			// 		if e.Error() == g.Error() {
			// 			found = true
			// 			g.Internal = nil
			// 			assert.EqualValues(t, e, g)
			// 			break
			// 		}
			// 	}
			//
			// 	if found {
			// 		continue
			// 	}
			//
			// 	require.True(t, found, "%+v", g.Internal.Description())
			// }
		})
	}
}
