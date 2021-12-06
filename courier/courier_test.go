package courier_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ory/kratos/x"
	gomail "github.com/ory/mail/v3"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	dhelper "github.com/ory/x/sqlcon/dockertest"

	courier "github.com/ory/kratos/courier"
	templates "github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
)

// nolint:staticcheck
func TestMain(m *testing.M) {
	atexit := dhelper.NewOnExit()
	atexit.Add(x.CleanUpTestSMTP)
	atexit.Exit(m.Run())
}

func TestNewSMTP(t *testing.T) {
	ctx := context.Background()

	setupConfig := func(stringURL string) *courier.Courier {
		conf, reg := internal.NewFastRegistryWithMocks(t)
		conf.MustSet(config.ViperKeyCourierSMTPURL, stringURL)
		t.Logf("SMTP URL: %s", conf.CourierSMTPURL().String())
		return courier.NewSMTP(ctx, reg)
	}

	if testing.Short() {
		t.SkipNow()
	}

	//Should enforce StartTLS => dialer.StartTLSPolicy = gomail.MandatoryStartTLS and dialer.SSL = false
	smtp := setupConfig("smtp://foo:bar@my-server:1234/")
	assert.Equal(t, smtp.Dialer.StartTLSPolicy, gomail.MandatoryStartTLS, "StartTLS not enforced")
	assert.Equal(t, smtp.Dialer.SSL, false, "Implicit TLS should not be enabled")

	//Should enforce TLS => dialer.SSL = true
	smtp = setupConfig("smtps://foo:bar@my-server:1234/")
	assert.Equal(t, smtp.Dialer.SSL, true, "Implicit TLS should be enabled")

	//Should allow cleartext => dialer.StartTLSPolicy = gomail.OpportunisticStartTLS and dialer.SSL = false
	smtp = setupConfig("smtp://foo:bar@my-server:1234/?disable_starttls=true")
	assert.Equal(t, smtp.Dialer.StartTLSPolicy, gomail.OpportunisticStartTLS, "StartTLS is enforced")
	assert.Equal(t, smtp.Dialer.SSL, false, "Implicit TLS should not be enabled")
}

func TestSMTP(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	smtp, api, err := x.RunTestSMTP()
	require.NoError(t, err)
	t.Logf("SMTP URL: %s", smtp)
	t.Logf("API URL: %s", api)

	ctx := context.Background()

	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeyCourierSMTPURL, smtp)
	conf.MustSet(config.ViperKeyCourierSMTPFrom, "test-stub@ory.sh")
	reg.Logger().Level = logrus.TraceLevel

	c := reg.Courier(ctx)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	id, err := c.QueueEmail(ctx, templates.NewTestStub(conf, &templates.TestStubModel{
		To:      "test-recipient-1@example.org",
		Subject: "test-subject-1",
		Body:    "test-body-1",
	}))
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)

	id, err = c.QueueEmail(ctx, templates.NewTestStub(conf, &templates.TestStubModel{
		To:      "test-recipient-2@example.org",
		Subject: "test-subject-2",
		Body:    "test-body-2",
	}))
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)

	// The third email contains a sender name and custom headers
	conf.MustSet(config.ViperKeyCourierSMTPFromName, "Bob")
	conf.MustSet(config.ViperKeyCourierSMTPHeaders+".test-stub-header1", "foo")
	conf.MustSet(config.ViperKeyCourierSMTPHeaders+".test-stub-header2", "bar")
	customerHeaders := conf.CourierSMTPHeaders()
	require.Len(t, customerHeaders, 2)
	id, err = c.QueueEmail(ctx, templates.NewTestStub(conf, &templates.TestStubModel{
		To:      "test-recipient-3@example.org",
		Subject: "test-subject-3",
		Body:    "test-body-3",
	}))
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)

	go func() {
		require.NoError(t, c.Work(ctx))
	}()

	var body []byte
	for k := 0; k < 30; k++ {
		time.Sleep(time.Second)
		err = func() error {
			res, err := http.Get(api + "/api/v2/messages")
			if err != nil {
				return err
			}

			defer res.Body.Close()
			body, err = ioutil.ReadAll(res.Body)
			if err != nil {
				return err
			}

			if http.StatusOK != res.StatusCode {
				return errors.Errorf("expected status code 200 but got %d with body: %s", res.StatusCode, body)
			}

			if total := gjson.GetBytes(body, "total").Int(); total != 3 {
				return errors.Errorf("expected to have delivered at least 3 messages but got count %d with body: %s", total, body)
			}

			return nil
		}()
		if err == nil {
			break
		}
	}
	require.NoError(t, err)

	for k := 1; k <= 3; k++ {
		assert.Contains(t, string(body), fmt.Sprintf("test-subject-%d", k))
		assert.Contains(t, string(body), fmt.Sprintf("test-body-%d", k))
		assert.Contains(t, string(body), fmt.Sprintf("test-recipient-%d@example.org", k))
		assert.Contains(t, string(body), "test-stub@ory.sh")
	}

	// Assertion for the third email with sender name and headers
	assert.Contains(t, string(body), "Bob")
	assert.Contains(t, string(body), `"test-stub-header1":["foo"]`)
	assert.Contains(t, string(body), `"test-stub-header2":["bar"]`)
}
