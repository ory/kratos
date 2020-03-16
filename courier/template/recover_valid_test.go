package template_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/internal"
)

func TestRecoverValid(t *testing.T) {
	conf, _ := internal.NewRegistryDefault(t)
	tpl := template.NewRecoverValid(conf, &template.RecoverValidModel{})

	rendered, err := tpl.EmailBody()
	require.NoError(t, err)
	assert.NotEmpty(t, rendered)

	rendered, err = tpl.EmailSubject()
	require.NoError(t, err)
	assert.NotEmpty(t, rendered)
}
