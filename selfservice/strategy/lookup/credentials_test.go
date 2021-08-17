package lookup

import (
	_ "embed"
	"encoding/json"
	"github.com/ory/x/assertx"
	"github.com/ory/x/sqlxx"
	"testing"
	"time"
)

//go:embed fixtures/node.json
var fixtureNode []byte

func TestToNode(t *testing.T) {
	c := CredentialsConfig{RecoveryCodes: []RecoveryCode{
		{Code: "foo", UsedAt: sqlxx.NullTime(time.Unix(1629199958, 0))},
		{Code: "bar"},
		{Code: "baz"},
		{Code: "oof", UsedAt: sqlxx.NullTime(time.Unix(1629199968, 0))},
	}}

	assertx.EqualAsJSON(t, json.RawMessage(fixtureNode), c.ToNode())
}
