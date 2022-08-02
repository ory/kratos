package courier_test

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier"
	templates "github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
)

func TestMessageRetries(t *testing.T) {
	ctx := context.Background()

	conf, reg := internal.NewRegistryDefaultWithDSN(t, "")
	conf.MustSet(config.ViperKeyCourierMessageRetries, 1)

	reg.Logger().Level = logrus.TraceLevel

	c := reg.Courier(ctx)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	id, err := c.QueueEmail(ctx, templates.NewTestStub(reg, &templates.TestStubModel{
		To:      "test-recipient-1@example.org",
		Subject: "test-subject-1",
		Body:    "test-body-1",
	}))
	require.NoError(t, err)
	require.NotZero(t, id)

	err = c.DispatchQueue(ctx)
	require.Error(t, err)

	err = c.DispatchQueue(ctx)
	require.NoError(t, err)

	var message courier.Message
	err = reg.Persister().GetConnection(ctx).
		Where("status = ?", courier.MessageStatusAbandoned).
		First(&message)

	require.NoError(t, err)
	require.Equal(t, id, message.ID)
}
