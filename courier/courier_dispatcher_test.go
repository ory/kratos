package courier_test

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier"
	templates "github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
)

func TestMessageTTL(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := context.Background()

	conf, reg := internal.NewRegistryDefaultWithDSN(t, "")
	conf.MustSet(config.ViperKeyCourierMessageTTL, 1*time.Nanosecond)

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
	require.NotEqual(t, uuid.Nil, id)

	c.DispatchQueue(ctx)

	time.Sleep(1 * time.Second)

	var message courier.Message
	err = reg.Persister().GetConnection(ctx).
		Where("status = ?", courier.MessageStatusAbandoned).
		First(&message)

	require.NoError(t, err)
	require.Equal(t, id, message.ID)
}
