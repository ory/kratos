package code_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/x/urlx"
)

func TestManager(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/default.schema.json")
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://www.ory.sh/")
	conf.MustSet(ctx, config.ViperKeyCourierSMTPURL, "smtp://foo@bar@dev.null/")
	conf.MustSet(ctx, config.ViperKeyLinkBaseURL, "https://link-url/")

	u := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}

	i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Traits = identity.Traits(`{"email": "tracked@ory.sh"}`)
	require.NoError(t, reg.IdentityManager().Create(context.Background(), i))

	hr := httptest.NewRequest("GET", "https://www.ory.sh", nil)

	t.Run("method=SendRecoveryCode", func(t *testing.T) {
		f, err := recovery.NewFlow(conf, time.Hour, "", u, reg.RecoveryStrategies(context.Background()), flow.TypeBrowser)
		require.NoError(t, err)

		require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(context.Background(), f))

		require.NoError(t, reg.RecoveryCodeSender().SendRecoveryCode(context.Background(), hr, f, "email", "tracked@ory.sh"))
		require.EqualError(t, reg.RecoveryCodeSender().SendRecoveryCode(context.Background(), hr, f, "email", "not-tracked@ory.sh"), link.ErrUnknownAddress.Error())

		messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
		require.NoError(t, err)
		require.Len(t, messages, 2)

		assert.EqualValues(t, "tracked@ory.sh", messages[0].Recipient)
		assert.Contains(t, messages[0].Subject, "Recover access to your account")

		assert.Regexp(t, `(\d{8})`, messages[0].Body)

		assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
		assert.Contains(t, messages[1].Subject, "Account access attempted")
		assert.NotContains(t, messages[1].Body, urlx.AppendPaths(conf.SelfServiceLinkMethodBaseURL(ctx), recovery.RouteSubmitFlow).String()+"?")
		assert.NotContains(t, messages[1].Body, "code=") // TODO: Might be wrong?
		assert.NotContains(t, messages[1].Body, "flow=")
	})

}
