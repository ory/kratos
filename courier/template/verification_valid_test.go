package template_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/internal"
)

func TestVerifyValid(t *testing.T) {
	conf, _ := internal.NewFastRegistryWithMocks(t)
	tpl := template.NewVerificationValid(conf, &template.VerificationValidModel{})

	rendered, err := tpl.EmailBody()
	require.NoError(t, err)
	assert.NotEmpty(t, rendered)

	rendered, err = tpl.EmailSubject()
	require.NoError(t, err)
	assert.NotEmpty(t, rendered)
}
