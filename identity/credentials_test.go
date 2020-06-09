package identity

import (
	"testing"

	"github.com/mohae/deepcopy"
	"github.com/stretchr/testify/assert"

	"github.com/ory/x/sqlxx"
)

func TestCredentialsEqual(t *testing.T) {
	original := map[CredentialsType]Credentials{
		"foo": {Type: "foo", Identifiers: []string{"bar"}, Config: sqlxx.JSONRawMessage(`{"foo":"bar"}`)},
	}

	derived := deepcopy.Copy(original).(map[CredentialsType]Credentials)
	assert.EqualValues(t, original, derived)
	derived["foo"].Identifiers[0] = "baz"
	assert.NotEqual(t, original, derived)
}
