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

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	dhelper "github.com/ory/x/sqlcon/dockertest"

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

	// The third email contains a sender name
	conf.MustSet(config.ViperKeyCourierSMTPFromName, "Bob")
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

	// Assertion for the third email with sender name
	assert.Contains(t, string(body), "Bob")
}
