// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template/sms"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
)

func TestQueueSMS(t *testing.T) {
	ctx := context.Background()

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

		rb, err := io.ReadAll(r.Body)
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
	t.Cleanup(srv.Close)

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
	conf.MustSet(ctx, config.ViperKeyCourierChannels, fmt.Sprintf(`[
		{
			"id": "sms",
			"type": "http",
			"request_config": %s
		}
	]`, requestConfig))
	conf.MustSet(ctx, config.ViperKeyCourierSMTPURL, "http://foo.url")
	reg.Logger().Level = logrus.TraceLevel

	c, err := reg.Courier(ctx)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(ctx)
	defer t.Cleanup(cancel)

	for _, message := range expectedSMS {
		id, err := c.QueueSMS(ctx, sms.NewTestStub(reg, message))
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, id)
	}

	require.NoError(t, c.DispatchQueue(ctx))

	require.Eventually(t, func() bool {
		return len(actual) == len(expectedSMS)
	}, 10*time.Second, 250*time.Millisecond)

	for i, message := range actual {
		expected := expectedSMS[i]

		assert.Equal(t, expected.To, message.To)
		assert.Equal(t, fmt.Sprintf("stub sms body %s\n", expected.Body), message.Body)
	}
}

func TestDisallowedInternalNetwork(t *testing.T) {
	ctx := context.Background()

	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeyCourierChannels, `[
		{
			"id": "sms",
			"type": "http",
			"request_config": {
				"url": "http://127.0.0.1/",
				"method": "GET",
				"body": "file://./stub/request.config.twilio.jsonnet"
			}
		}
	]`)
	conf.MustSet(ctx, config.ViperKeyCourierSMTPURL, "http://foo.url")
	conf.MustSet(ctx, config.ViperKeyClientHTTPNoPrivateIPRanges, true)

	c, err := reg.Courier(ctx)
	require.NoError(t, err)
	c.(interface {
		FailOnDispatchError()
	}).FailOnDispatchError()
	_, err = c.QueueSMS(ctx, sms.NewTestStub(reg, &sms.TestStubModel{
		To:   "+12065550101",
		Body: "test-sms-body-1",
	}))
	require.NoError(t, err)

	err = c.DispatchQueue(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not a permitted destination")
}
