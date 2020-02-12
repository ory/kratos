package identity

import (
	"encoding/json"
	"testing"

	"github.com/mohae/deepcopy"
	"github.com/stretchr/testify/assert"
)

func TestCredentialsEqual(t *testing.T) {
	original := map[CredentialsType]Credentials{
		"foo": {Type: "foo", Identifiers: []string{"bar"}, Config: json.RawMessage(`{"foo":"bar"}`)},
	}

	derived := deepcopy.Copy(original).(map[CredentialsType]Credentials)
	assert.EqualValues(t, original, derived)
	derived["foo"].Identifiers[0] = "baz"
	assert.NotEqual(t, original, derived)
}
