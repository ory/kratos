package otp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/selfservice/strategy/otp"
	"github.com/ory/x/urlx"
)

func TestSender(t *testing.T) {
	ctx := context.Background()
	u := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}
	hr := httptest.NewRequest("GET", "https://www.ory.sh", nil)

	t.Run("send OTP via phone", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/phone.schema.json")

		conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://www.ory.sh/")
		conf.MustSet(ctx, config.ViperKeyCourierSMSRequestConfig, `  sms:
    enabled: true
    from: '+49123456789'
    request_config:
      url: https://api.twilio.com/2010-04-01/Accounts/YourAccountID/Messages.json
      method: POST
      body: base64://e30=
      header:
        'Content-Type': 'application/x-www-form-urlencoded'
      auth:
        type: basic_auth
        config:
          user: YourUsername
          password: YourPass`)
		conf.MustSet(ctx, config.ViperKeyLinkBaseURL, "https://link-url/")

		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		i.Traits = identity.Traits(`{"phone": "+18004444444"}`)
		require.NoError(t, reg.IdentityManager().Create(ctx, i))

		t.Run("case=recovery", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, time.Hour, "", u, reg.RecoveryStrategies(ctx), flow.TypeBrowser)
			require.NoError(t, err)

			require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(ctx, f))

			require.EqualError(t, reg.OTPSender().SendRecoveryOTP(ctx, hr, f, "+380634872774"), otp.ErrUnknownIdentifier.Error())
			require.NoError(t, reg.OTPSender().SendRecoveryOTP(ctx, hr, f, "+18004444444"))

			messages, err := reg.CourierPersister().NextMessages(ctx, 12)
			require.NoError(t, err)
			require.Len(t, messages, 1)

			assert.EqualValues(t, "+18004444444", messages[0].Recipient)
			assert.Regexp(t, regexp.MustCompile(`"Code":"\w{8}"`), string(messages[0].TemplateData))
		})

		t.Run("case=verification", func(t *testing.T) {
			f, err := verification.NewFlow(conf, time.Hour, "", u, reg.VerificationStrategies(ctx), flow.TypeBrowser)
			require.NoError(t, err)

			require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

			require.EqualError(t, reg.OTPSender().SendVerificationOTP(ctx, f, "+380634872775"), otp.ErrUnknownIdentifier.Error())
			require.NoError(t, reg.OTPSender().SendVerificationOTP(ctx, f, "+18004444444"))

			messages, err := reg.CourierPersister().NextMessages(ctx, 12)
			require.NoError(t, err)
			require.Len(t, messages, 1)

			assert.EqualValues(t, "+18004444444", messages[0].Recipient)
			assert.Regexp(t, regexp.MustCompile(`"Code":"\w{8}"`), string(messages[0].TemplateData))

			address, err := reg.IdentityPool().FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypePhone, "+18004444444")
			require.NoError(t, err)
			assert.EqualValues(t, identity.VerifiableAddressStatusSent, address.Status)
		})
	})

	t.Run("send OTP via email", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)

		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/email.schema.json")
		conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://www.ory.sh/")
		conf.MustSet(ctx, config.ViperKeyCourierSMTPURL, "smtp://foo@bar@dev.null/")
		conf.MustSet(ctx, config.ViperKeyLinkBaseURL, "https://link-url/")

		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		i.Traits = identity.Traits(`{"email": "tracked@ory.sh"}`)
		require.NoError(t, reg.IdentityManager().Create(ctx, i))

		t.Run("case=recovery", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, time.Hour, "", u, reg.RecoveryStrategies(ctx), flow.TypeBrowser)
			require.NoError(t, err)

			require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(ctx, f))

			require.NoError(t, reg.OTPSender().SendRecoveryOTP(ctx, hr, f, "tracked@ory.sh"))
			require.EqualError(t, reg.OTPSender().SendRecoveryOTP(ctx, hr, f, "not-tracked@ory.sh"), otp.ErrUnknownIdentifier.Error())

			messages, err := reg.CourierPersister().NextMessages(ctx, 12)
			require.NoError(t, err)
			require.Len(t, messages, 2)

			assert.EqualValues(t, "tracked@ory.sh", messages[0].Recipient)
			assert.Contains(t, messages[0].Subject, "Recover access to your account")
			assert.Regexp(t, regexp.MustCompile(`please recover access to your account by entering following code:\n\n(\w{8})`), messages[0].Body)

			assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
			assert.Contains(t, messages[1].Subject, "Account access attempted")
			assert.Regexp(t, regexp.MustCompile(`please recover access to your account by entering following code:\n\n(\w{8})`), messages[0].Body)
		})

		t.Run("case=verification", func(t *testing.T) {
			f, err := verification.NewFlow(conf, time.Hour, "", u, reg.VerificationStrategies(ctx), flow.TypeBrowser)
			require.NoError(t, err)

			require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

			require.NoError(t, reg.LinkSender().SendVerificationLink(ctx, f, "email", "tracked@ory.sh"))
			require.EqualError(t, reg.LinkSender().SendVerificationLink(ctx, f, "email", "not-tracked@ory.sh"), link.ErrUnknownAddress.Error())
			messages, err := reg.CourierPersister().NextMessages(ctx, 12)
			require.NoError(t, err)
			require.Len(t, messages, 2)

			assert.EqualValues(t, "tracked@ory.sh", messages[0].Recipient)
			assert.Contains(t, messages[0].Subject, "Please verify")
			assert.Contains(t, messages[0].Body, urlx.AppendPaths(conf.SelfServiceLinkMethodBaseURL(ctx), verification.RouteSubmitFlow).String()+"?")
			assert.Contains(t, messages[0].Body, "token=")
			assert.Contains(t, messages[0].Body, "flow=")

			assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
			assert.Contains(t, messages[1].Subject, "tried to verify")
			assert.NotContains(t, messages[1].Body, urlx.AppendPaths(conf.SelfServiceLinkMethodBaseURL(ctx), verification.RouteSubmitFlow).String()+"?")

			address, err := reg.IdentityPool().FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, "tracked@ory.sh")
			require.NoError(t, err)
			assert.EqualValues(t, identity.VerifiableAddressStatusSent, address.Status)
		})
	})
}
