package x

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ConvertibleBooleanTest struct {
	Verified ConvertibleBoolean `json:"verified,omitempty"`
}

func TestUnmarshalBool(t *testing.T) {
	data := `{"verified":true}`
	c := ConvertibleBooleanTest{}
	err := json.Unmarshal([]byte(data), &c)
	require.NoError(t, err)
	assert.Equal(t, ConvertibleBoolean(true), c.Verified)

	data = `{"verified":false}`
	err = json.Unmarshal([]byte(data), &c)
	require.NoError(t, err)
	assert.Equal(t, ConvertibleBoolean(false), c.Verified)
}

func TestUnmarshalString(t *testing.T) {
	data := `{"verified":"true"}`
	c := ConvertibleBooleanTest{}
	err := json.Unmarshal([]byte(data), &c)
	require.NoError(t, err)
	assert.Equal(t, ConvertibleBoolean(true), c.Verified)

	data = `{"verified":"false"}`
	err = json.Unmarshal([]byte(data), &c)
	require.NoError(t, err)
	assert.Equal(t, ConvertibleBoolean(false), c.Verified)
}

func TestUnmarshalError(t *testing.T) {
	data := `{"verified":1}`
	c := ConvertibleBooleanTest{}
	err := json.Unmarshal([]byte(data), &c)
	assert.Error(t, err)
}
