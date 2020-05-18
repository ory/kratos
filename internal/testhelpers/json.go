package testhelpers

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func JSONEq(t *testing.T, expected, actual interface{}, messageAndArgs ...interface{}) {
	var eb, ab bytes.Buffer
	require.NoError(t, json.NewEncoder(&eb).Encode(expected))
	require.NoError(t, json.NewEncoder(&ab).Encode(actual))
	assert.JSONEq(t, eb.String(), ab.String(), messageAndArgs...)
}
