package testhelpers

import (
	"context"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupRemoteConfig(t *testing.T, ctx context.Context, plaintext string, html string, subject string) *driver.RegistryDefault {
	_, reg := internal.NewFastRegistryWithMocks(t)
	require.NoError(t, reg.Config(ctx).Set(config.ViperKeyCourierTemplatesRecoveryInvalid, &config.CourierEmailTemplate{
		TemplateRoot: "",
		Body: &config.CourierEmailBodyTemplate{
			PlainText: plaintext,
			HTML:      html,
		},
		Subject: subject,
	}))
	return reg
}

func TestRendered(t *testing.T, ctx context.Context, tpl interface {
	EmailBody(context.Context) (string, error)
	EmailSubject(context.Context) (string, error)
}) {
	rendered, err := tpl.EmailBody(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, rendered)

	rendered, err = tpl.EmailSubject(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, rendered)
}
