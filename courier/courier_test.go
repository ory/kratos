package courier_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/dockertest"

	"github.com/ory/viper"
	dhelper "github.com/ory/x/sqlcon/dockertest"

	templates "github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/internal"
)

var resources []*dockertest.Resource

// nolint:staticcheck
func TestMain(m *testing.M) {
	atexit := dhelper.NewOnExit()
	atexit.Add(func() {
		for _, resource := range resources {
			resource.Close()
		}
	})
	atexit.Exit(m.Run())
}

func runTestSMTP(t *testing.T) (smtp, api string) {
	if smtp, api := os.Getenv("TEST_MAILHOG_SMTP"), os.Getenv("TEST_MAILHOG_API"); smtp != "" && api != "" {
		t.Logf("Skipping Docker setup because environment variables TEST_MAILHOG_SMTP and TEST_MAILHOG_API are both set.")
		return smtp, api
	} else if len(smtp)+len(api) > 0 {
		t.Fatal("Environment variables TEST_MAILHOG_SMTP, TEST_MAILHOG_API must both be set!")
		return "", ""
	}

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := pool.
		RunWithOptions(&dockertest.RunOptions{
			Repository: "mailhog/mailhog",
			Tag:        "v1.0.0",
			Cmd: []string{
				"-invite-jim",
				"-jim-linkspeed-affect=0.05",
				"-jim-reject-auth=0.05",
				"-jim-reject-recipient=0.05",
				"-jim-reject-sender=0.05",
				"-jim-disconnect=0.05",
				"-jim-linkspeed-min=1250",
				"-jim-linkspeed-max=12500",
			},
		})
	require.NoError(t, err)
	resources = append(resources, resource)

	smtp = fmt.Sprintf("smtp://test:test@127.0.0.1:%s", resource.GetPort("1025/tcp"))
	api = fmt.Sprintf("http://127.0.0.1:%s", resource.GetPort("8025/tcp"))
	require.NoError(t, backoff.Retry(func() error {
		res, err := http.Get(api + "/api/v2/messages")
		if err != nil {
			t.Logf("Unable to connect to mailhog: %s", err)
			return err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			err := errors.Errorf("expected status code 200 but got: %d", res.StatusCode)
			t.Logf("Unable to connect to mailhog: %s", err)
			return err
		}
		return nil
	}, backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 15)))

	return smtp, api
}

func TestSMTP(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	smtp, api := runTestSMTP(t)

	conf, reg := internal.NewRegistryDefault(t)
	viper.Set(configuration.ViperKeyCourierSMTPURL, smtp)
	viper.Set(configuration.ViperKeyCourierSMTPFrom, "test-stub@ory.sh")
	c := reg.Courier()

	go func() {
		require.NoError(t, c.Work(context.Background()))
	}()

	t.Run("case=queue messages", func(t *testing.T) {
		id, err := c.SendEmail(context.Background(), templates.NewTestStub(conf, &templates.TestStubModel{
			To:      "test-recipient-1@example.org",
			Subject: "test-subject-1",
			Body:    "test-body-1",
		}))
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, id)

		id, err = c.SendEmail(context.Background(), templates.NewTestStub(conf, &templates.TestStubModel{
			To:      "test-recipient-2@example.org",
			Subject: "test-subject-2",
			Body:    "test-body-2",
		}))
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, id)
	})

	t.Run("case=check for delivery", func(t *testing.T) {
		var err error
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

				if total := gjson.GetBytes(body, "total").Int(); total != 2 {
					return errors.Errorf("expected to have delivered at least 2 messages but got count %d with body: %s", total, body)
				}

				return nil
			}()
			if err == nil {
				break
			}
		}
		require.NoError(t, err)

		for k := 1; k <= 2; k++ {
			assert.Contains(t, string(body), fmt.Sprintf("test-subject-%d", k))
			assert.Contains(t, string(body), fmt.Sprintf("test-body-%d", k))
			assert.Contains(t, string(body), fmt.Sprintf("test-recipient-%d@example.org", k))
			assert.Contains(t, string(body), "test-stub@ory.sh")
		}
	})
}
