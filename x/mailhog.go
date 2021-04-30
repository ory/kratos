package x

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"

	"github.com/ory/dockertest/v3"
)

var resources []*dockertest.Resource

func CleanUpTestSMTP() {
	for _, resource := range resources {
		resource.Close()
	}
}

func RunTestSMTP() (smtp, api string, err error) {
	if smtp, api := os.Getenv("TEST_MAILHOG_SMTP"), os.Getenv("TEST_MAILHOG_API"); smtp != "" && api != "" {
		return smtp, api, nil
	} else if len(smtp)+len(api) > 0 {
		return "", "", errors.New("environment variables TEST_MAILHOG_SMTP, TEST_MAILHOG_API must both be set!")
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", "", err
	}

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
	if err != nil {
		return "", "", err
	}
	resources = append(resources, resource)

	smtp = fmt.Sprintf("smtp://test:test@127.0.0.1:%s", resource.GetPort("1025/tcp"))
	api = fmt.Sprintf("http://127.0.0.1:%s", resource.GetPort("8025/tcp"))
	if err := backoff.Retry(func() error {
		res, err := http.Get(api + "/api/v2/messages")
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			err := errors.Errorf("expected status code 200 but got: %d", res.StatusCode)
			return err
		}
		return nil
	}, backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 15)); err != nil {
		return "", "", err
	}

	return smtp, api, nil
}
