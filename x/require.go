package x

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func RequireJSONMarshal(t *testing.T, in interface{}) []byte {
	var b bytes.Buffer
	require.NoError(t, json.NewEncoder(&b).Encode(in))
	return b.Bytes()
}

func RequireNotNilUUID(t *testing.T, id uuid.UUID) {
	require.NotEqual(t, uuid.Nil, id)
}
