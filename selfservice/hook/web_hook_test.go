package hook

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"io"
	"strings"
	"testing"
)

type testData struct {
	ID uuid.UUID `json:"id"`
	Foo string `json:"foo"`
}

func TestJsonNet(t *testing.T) {
	id, _ := uuid.NewV1()
	td := testData{id, "Bar"}
	h := NewWebHook(nil, json.RawMessage{})

	b, err := h.createBody("test_body.jsonnet", &td, &td)
	require.NoError(t, err)

	buf := new(strings.Builder)
	io.Copy(buf, b)

	expected := fmt.Sprintf("{\n	\"flow_id\":\"{%s}\",\"session_id\":\"{%s}\"}", td.ID, td.Foo)

	require.Equal(t, buf.String(), expected)
}
