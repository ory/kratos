package x

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeMap(t *testing.T) {
	r, err := TypeMap(map[string]string{
		"string":  "string",
		"int":     "123",
		"float":   "123.123",
		"bool":    "TrUe",
		"bool_on": "oN",
	})
	require.NoError(t, err)

	assert.Equal(t, map[string]interface{}{
		"string":  "string",
		"int":     int64(123),
		"float":   123.123,
		"bool":    true,
		"bool_on": true,
	}, r)

	_, err = TypeMap(map[string]string{
		"int": "999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999",
	})
	assert.Error(t, err)

	_, err = TypeMap(map[string]string{
		"float": "999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999.9",
	})
	assert.Error(t, err)
}

func TestUntypedMapToJSON(t *testing.T) {
	r, err := UntypedMapToJSON(map[string]string{
		"string":  "string",
		"int":     "123",
		"float":   "123.123",
		"bool":    "TrUe",
		"bool_on": "oN",
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{"string":"string","int":123,"float":123.123,"bool":true,"bool_on":true}`, string(r))

	_, err = UntypedMapToJSON(map[string]string{
		"int": "999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999",
	})
	assert.Error(t, err)
}
