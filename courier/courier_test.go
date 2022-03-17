package courier_test

import (
	"testing"

	"github.com/ory/kratos/x"
	dhelper "github.com/ory/x/sqlcon/dockertest"
)

// nolint:staticcheck
func TestMain(m *testing.M) {
	atexit := dhelper.NewOnExit()
	atexit.Add(x.CleanUpTestSMTP)
	atexit.Exit(m.Run())
}
