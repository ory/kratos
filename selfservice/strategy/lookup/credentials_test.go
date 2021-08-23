package lookup

import (
	_ "embed"
	"encoding/json"
	"testing"
	"time"

	"github.com/ory/x/assertx"
	"github.com/ory/x/sqlxx"
)

//go:embed fixtures/node.json
var fixtureNode []byte

func TestToNode(t *testing.T) {
	c := CredentialsConfig{RecoveryCodes: []RecoveryCode{
		{Code: "foo", UsedAt: sqlxx.NullTime(time.Unix(1629199958, 0).UTC())},
		{Code: "bar"},
		{Code: "baz"},
		{Code: "oof", UsedAt: sqlxx.NullTime(time.Unix(1629199968, 0).UTC())},
	}}

	assertx.EqualAsJSON(t, json.RawMessage(fixtureNode), c.ToNode())
}
