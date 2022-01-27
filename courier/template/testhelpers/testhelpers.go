package testhelpers

import (
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupRemoteConfig(t *testing.T, plaintext string, html string, subject string) *config.Config {
	conf, _ := internal.NewFastRegistryWithMocks(t)
	require.NoError(t, conf.Set(config.ViperKeyCourierTemplatesRecoveryInvalid, &config.CourierEmailTemplate{
		TemplateRoot: "",
		Body: &config.CourierEmailBodyTemplate{
			PlainText: plaintext,
			HTML:      html,
		},
		Subject: subject,
	}))
	return conf
}

func TestRendered(t *testing.T, tpl interface {
	EmailBody() (string, error)
	EmailSubject() (string, error)
}) {
	rendered, err := tpl.EmailBody()
	require.NoError(t, err)
	assert.NotEmpty(t, rendered)

	rendered, err = tpl.EmailSubject()
	require.NoError(t, err)
	assert.NotEmpty(t, rendered)
}
