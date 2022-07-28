package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/x"
)

func TestGenerateState(t *testing.T) {
	state := generateState(x.NewUUID().String())
	assert.NotEmpty(t, state)
	t.Logf("state: %s", state)
}
