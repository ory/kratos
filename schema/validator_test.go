package schema

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/stringsx"
)

func TestSchemaValidator(t *testing.T) {
	router := httprouter.New()
	fs := http.StripPrefix("/schema", http.FileServer(http.Dir("stub/validator")))
	router.GET("/schema/:name", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fs.ServeHTTP(w, r)
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
			err: "I[#/age] S[#/properties/age/minimum] must be >= 1 but found -1",
		},
		{
			i:   json.RawMessage(`{ "whatever": "first-name", "lastName": "last-name", "age": 1 }`),
			err: `I[#] S[#/additionalProperties] additionalProperties "whatever" not allowed`,
		},
		{
			u: ts.URL + "/schema/whatever.schema.json",
			i: json.RawMessage(`{ "whatever": "first-name", "lastName": "last-name", "age": 1 }`),
		},
		{
			u:   ts.URL + "/schema/whatever.schema.json",
			i:   json.RawMessage(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			err: `I[#] S[#/additionalProperties] additionalProperties "firstName" not allowed`,
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
			err := NewValidator().Validate(stringsx.Coalesce(tc.u, ts.URL+"/schema/firstName.schema.json"), tc.i, )
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.err)
			}
		})
	}
}
