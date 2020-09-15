package link_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/link"
)

func TestManager(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/default.schema.json")
	viper.Set(configuration.ViperKeyPublicBaseURL, "https://www.ory.sh/")
	viper.Set(configuration.ViperKeyCourierSMTPURL, "smtp://foo@bar@dev.null/")

	u := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}

	i := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
	i.Traits = identity.Traits(`{"email": "tracked@ory.sh"}`)
	require.NoError(t, reg.IdentityManager().Create(context.Background(), i))

	t.Run("method=SendRecoveryLink", func(t *testing.T) {
		f, err := recovery.NewFlow(time.Hour, "", u, reg.RecoveryStrategies(), flow.TypeBrowser)
		require.NoError(t, err)

		require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(context.Background(), f))

		require.NoError(t, reg.LinkSender().SendRecoveryLink(context.Background(), f, "email", "tracked@ory.sh"))
		require.EqualError(t, reg.LinkSender().SendRecoveryLink(context.Background(), f, "email", "not-tracked@ory.sh"), link.ErrUnknownAddress.Error())

		messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
		require.NoError(t, err)
		require.Len(t, messages, 2)

		assert.EqualValues(t, "tracked@ory.sh", messages[0].Recipient)
		assert.Contains(t, messages[0].Subject, "Recover access to your account")
		assert.Contains(t, messages[0].Body, urlx.AppendPaths(conf.SelfPublicURL(), link.RouteRecovery).String()+"?token=")

		assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
		assert.Contains(t, messages[1].Subject, "Account access attempted")
		assert.NotContains(t, messages[1].Body, urlx.AppendPaths(conf.SelfPublicURL(), link.RouteRecovery).String()+"?token=")
	})

	t.Run("method=SendVerificationLink", func(t *testing.T) {
		f, err := verification.NewFlow(time.Hour, "", u, reg.VerificationStrategies(), flow.TypeBrowser)
		require.NoError(t, err)

		require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(context.Background(), f))

		require.NoError(t, reg.LinkSender().SendVerificationLink(context.Background(), f, "email", "tracked@ory.sh"))
		require.EqualError(t, reg.LinkSender().SendVerificationLink(context.Background(), f, "email", "not-tracked@ory.sh"), link.ErrUnknownAddress.Error())

		messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
		require.NoError(t, err)
		require.Len(t, messages, 4)

		assert.EqualValues(t, "tracked@ory.sh", messages[2].Recipient)
		assert.Contains(t, messages[2].Subject, "Please verify")
		assert.Contains(t, messages[2].Body, urlx.AppendPaths(conf.SelfPublicURL(), link.RouteVerification).String()+"?token=")

		assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
		assert.Contains(t, messages[3].Subject, "tried to verify")
		assert.NotContains(t, messages[3].Body, urlx.AppendPaths(conf.SelfPublicURL(), link.RouteVerification).String()+"?token=")
	})
}
