package lookup_test

import (
	_ "embed"
	"testing"
	"time"

	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/strategy/lookup"

	"github.com/ory/x/sqlxx"
)

func TestToNode(t *testing.T) {
	c := lookup.CredentialsConfig{RecoveryCodes: []lookup.RecoveryCode{
		{Code: "foo", UsedAt: sqlxx.NullTime(time.Unix(1629199958, 0).UTC())},
		{Code: "bar"},
		{Code: "baz"},
		{Code: "oof", UsedAt: sqlxx.NullTime(time.Unix(1629199968, 0).UTC())},
	}}

	testhelpers.SnapshotTExcept(t, c.ToNode(), []string{})
}
