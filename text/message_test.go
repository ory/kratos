package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage(t *testing.T) {
	expected := &Message{ID: InfoSelfServiceSettingsUpdateSuccess, Text: "foo", Type: Info}

	v, err := expected.Value()
	require.NoError(t, err)

	var actual Message
	require.NoError(t, actual.Scan(v.(string)))

	assert.EqualValues(t, expected, &actual, v)
}

func TestMessages(t *testing.T) {
	expected := Messages{{ID: InfoSelfServiceSettingsUpdateSuccess, Text: "foo", Type: Info}}

	v, err := expected.Value()
	require.NoError(t, err)

	var actual Messages
	require.NoError(t, actual.Scan(v.(string)))

	assert.EqualValues(t, expected, actual, v)
}
