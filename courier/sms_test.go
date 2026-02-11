// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template/sms"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/pkg"
	"github.com/ory/x/configx"
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

	actual := make(chan *sms.TestStubModel, len(expectedSMS))
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
		actual <- &sms.TestStubModel{
			To:   body.To,
			Body: body.Body,
		}
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

	_, reg := pkg.NewFastRegistryWithMocks(t, configx.WithValues(map[string]any{
		config.ViperKeyCourierChannels: fmt.Sprintf(`[{
			"id": "sms",
			"type": "http",
			"request_config": %s
		}]`, requestConfig),
		config.ViperKeyCourierSMTPURL: "http://foo.url",
	}))

	c, err := reg.Courier(t.Context())
	require.NoError(t, err)

	for _, message := range expectedSMS {
		id, err := c.QueueSMS(t.Context(), sms.NewTestStub(message))
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, id)
	}

	require.NoError(t, c.DispatchQueue(t.Context()))
	close(actual)

	require.Len(t, actual, len(expectedSMS))

	i := 0
	for message := range actual {
		expected := expectedSMS[i]

		assert.Equal(t, expected.To, message.To)
		assert.Equal(t, expected.Body, message.Body)
		i++
	}
}

func TestDisallowedInternalNetwork(t *testing.T) {
	_, reg := pkg.NewFastRegistryWithMocks(t, configx.WithValues(map[string]any{
		config.ViperKeyCourierChannels: `[
			{
				"id": "sms",
				"type": "http",
				"request_config": {
					"url": "http://127.0.0.1/",
					"method": "GET",
					"body": "file://./stub/request.config.twilio.jsonnet"
				}
			}
		]`,
		config.ViperKeyCourierSMTPURL:              "http://foo.url",
		config.ViperKeyClientHTTPNoPrivateIPRanges: true,
	}))

	c, err := reg.Courier(t.Context())
	require.NoError(t, err)
	c.(interface {
		FailOnDispatchError()
	}).FailOnDispatchError()
	_, err = c.QueueSMS(t.Context(), sms.NewTestStub(&sms.TestStubModel{
		To:   "+12065550101",
		Body: "test-sms-body-1",
	}))
	require.NoError(t, err)

	err = c.DispatchQueue(t.Context())
	assert.ErrorContains(t, err, "is not a permitted destination")
}
