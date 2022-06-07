package link_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/x/urlx"
)

func TestManager(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeyPublicBaseURL, "https://www.ory.sh/")
	conf.MustSet(config.ViperKeyCourierSMTPURL, "smtp://foo@bar@dev.null/")
	conf.MustSet(config.ViperKeyLinkBaseURL, "https://link-url/")
	conf.MustSet(config.ViperKeyIdentitySchemas, []config.Schema{
		{ID: "default", URL: "file://./stub/default.schema.json"},
		{ID: "phone", URL: "file://./stub/phone.schema.json"},
	})

	u := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}

	i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Traits = identity.Traits(`{"email": "tracked@ory.sh"}`)
	require.NoError(t, reg.IdentityManager().Create(context.Background(), i))

	i2 := identity.NewIdentity("phone")
	i2.Traits = identity.Traits(`{"phone": "+12345678901"}`)
	require.NoError(t, reg.IdentityManager().Create(context.Background(), i2))

	hr := httptest.NewRequest("GET", "https://www.ory.sh", nil)

	t.Run("method=SendRecoveryLink", func(t *testing.T) {
		f, err := recovery.NewFlow(conf, time.Hour, "", u, reg.RecoveryStrategies(context.Background()), flow.TypeBrowser)
		require.NoError(t, err)

		require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(context.Background(), f))

		require.NoError(t, reg.LinkSender().SendRecoveryLink(context.Background(), hr, f, "email", "tracked@ory.sh"))
		require.EqualError(t, reg.LinkSender().SendRecoveryLink(context.Background(), hr, f, "email", "not-tracked@ory.sh"), link.ErrUnknownAddress.Error())

		messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
		require.NoError(t, err)
		require.Len(t, messages, 2)

		assert.EqualValues(t, "tracked@ory.sh", messages[0].Recipient)
		assert.Contains(t, messages[0].Subject, "Recover access to your account")
		assert.Contains(t, messages[0].Body, urlx.AppendPaths(conf.SelfServiceLinkMethodBaseURL(), recovery.RouteSubmitFlow).String()+"?")
		assert.Contains(t, messages[0].Body, "token=")
		assert.Contains(t, messages[0].Body, "flow=")

		assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
		assert.Contains(t, messages[1].Subject, "Account access attempted")
		assert.NotContains(t, messages[1].Body, urlx.AppendPaths(conf.SelfServiceLinkMethodBaseURL(), recovery.RouteSubmitFlow).String()+"?")
		assert.NotContains(t, messages[1].Body, "token=")
		assert.NotContains(t, messages[1].Body, "flow=")
	})

	t.Run("method=SendVerificationLink for email", func(t *testing.T) {
		f, err := verification.NewFlow(conf, time.Hour, "", u, reg.VerificationStrategies(context.Background()), flow.TypeBrowser)
		require.NoError(t, err)

		require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(context.Background(), f))

		require.NoError(t, reg.LinkSender().SendVerificationLink(context.Background(), f, "email", "tracked@ory.sh"))
		require.EqualError(t, reg.LinkSender().SendVerificationLink(context.Background(), f, "email", "not-tracked@ory.sh"), link.ErrUnknownAddress.Error())
		messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
		require.NoError(t, err)
		require.Len(t, messages, 2)

		assert.EqualValues(t, "tracked@ory.sh", messages[0].Recipient)
		assert.Contains(t, messages[0].Subject, "Please verify")
		assert.Contains(t, messages[0].Body, urlx.AppendPaths(conf.SelfServiceLinkMethodBaseURL(), verification.RouteSubmitFlow).String()+"?")
		assert.Contains(t, messages[0].Body, "token=")
		assert.Contains(t, messages[0].Body, "flow=")

		assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
		assert.Contains(t, messages[1].Subject, "tried to verify")
		assert.NotContains(t, messages[1].Body, urlx.AppendPaths(conf.SelfServiceLinkMethodBaseURL(), verification.RouteSubmitFlow).String()+"?")
		address, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, "tracked@ory.sh")
		require.NoError(t, err)
		assert.EqualValues(t, identity.VerifiableAddressStatusSent, address.Status)
	})

	t.Run("method=SendVerificationLink for phone", func(t *testing.T) {
		f, err := verification.NewFlow(conf, time.Hour, "", u, reg.VerificationStrategies(context.Background()), flow.TypeBrowser)
		require.NoError(t, err)

		require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(context.Background(), f))

		require.NoError(t, reg.LinkSender().SendVerificationLink(context.Background(), f, "phone", "+12345678901"))
		require.EqualError(t, reg.LinkSender().SendVerificationLink(context.Background(), f, "phone", "+6789012345"), link.ErrUnknownAddress.Error())
		messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
		require.NoError(t, err)
		require.Len(t, messages, 1)

		assert.EqualValues(t, "+12345678901", messages[0].Recipient)
		assert.Contains(t, string(messages[0].TemplateData), urlx.AppendPaths(conf.SelfServiceLinkMethodBaseURL(), verification.RouteSubmitFlow).String()+"?")
		assert.Contains(t, string(messages[0].TemplateData), "token=")
		assert.Contains(t, string(messages[0].TemplateData), "flow=")
	})
}
