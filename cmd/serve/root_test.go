package serve_test

import (
	"testing"

	"github.com/ory/kratos/internal/testhelpers"
)

func TestServe(t *testing.T) {
	_, _ = testhelpers.StartE2EServer(t, "../../contrib/quickstart/kratos/email-password/kratos.yml", nil)
}
