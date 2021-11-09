package courier_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template/sms"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
)

func TestQueueSMS(t *testing.T) {
	expectedSender := "Kratos Test"
	expectedSMS := []*sms.TestStubModel{
		{
			To:   "+12065550101",
			Body: "test-sms-body-1",
		},
		{
			To:   "+12065550102",
			Body: "test-sms-body-2",
		},
	}

	actual := make([]*sms.TestStubModel, 0, 2)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type sendSMSRequestBody struct {
			To   string
			From string
			Body string
		}

		rb, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		var body sendSMSRequestBody

		err = json.Unmarshal(rb, &body)
		require.NoError(t, err)

		assert.NotEmpty(t, r.Header["Authorization"])
		assert.Equal(t, "Basic bWU6MTIzNDU=", r.Header["Authorization"][0])

		assert.Equal(t, body.From, expectedSender)
		actual = append(actual, &sms.TestStubModel{
			To:   body.To,
			Body: body.Body,
		})
	}))

	requestConfig := fmt.Sprintf(`{
		"url": "%s",
		"method": "POST",
		"body": "file://./stub/request.config.twilio.jsonnet",
		"auth": {
			"type": "basic_auth",
			"config": {
				"user":     "me",
				"password": "12345"
			}
		}
	}`, srv.URL)

	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeyCourierSMSRequestConfig, requestConfig)
	conf.MustSet(config.ViperKeyCourierSMSFrom, expectedSender)
	conf.MustSet(config.ViperKeyCourierSMSEnabled, true)
	conf.MustSet(config.ViperKeyCourierSMTPURL, "http://foo.url")
	reg.Logger().Level = logrus.TraceLevel

	ctx := context.Background()

	c := reg.Courier(ctx)

	ctx, cancel := context.WithCancel(ctx)
	defer t.Cleanup(cancel)

	for _, message := range expectedSMS {
		id, err := c.QueueSMS(ctx, sms.NewTestStub(reg, message))
		require.NoError(t, err)
		x.RequireNotNilUUID(t, id)
	}

	go func() {
		require.NoError(t, c.Work(ctx))
	}()

	time.Sleep(time.Second)
	for i, message := range actual {
		expected := expectedSMS[i]

		assert.Equal(t, expected.To, message.To)
		assert.Equal(t, fmt.Sprintf("stub sms body %s\n", expected.Body), message.Body)
	}

	srv.Close()
}
