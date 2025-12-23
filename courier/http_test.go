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

	"github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/x/configx"
)

func TestQueueHTTPEmail(t *testing.T) {
	type sendEmailRequestBody struct {
		IdentityID       string `json:"identity_id"`
		IdentityEmail    string `json:"identity_email"`
		Recipient        string `json:"recipient"`
		TemplateType     string `json:"template_type"`
		To               string `json:"to"`
		RecoveryCode     string `json:"recovery_code"`
		RecoveryURL      string `json:"recovery_url"`
		VerificationURL  string `json:"verification_url"`
		VerificationCode string `json:"verification_code"`
		Body             string `json:"body"`
		HTMLBody         string `json:"html_body"`
		Subject          string `json:"subject"`
	}

	expectedEmail := []*email.TestStubModel{
		{
			To:       "test-2@test.com",
			Subject:  "test-mailer-subject-1",
			Body:     "test-mailer-body-1",
			HTMLBody: "<html><body>test-mailer-body-html-1</body></html>",
		},
		{
			To:      "test-2@test.com",
			Subject: "test-mailer-subject-2",
			Body:    "test-mailer-body-2",
		},
	}

	actual := make(chan sendEmailRequestBody, len(expectedEmail))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rb, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var body sendEmailRequestBody

		err = json.Unmarshal(rb, &body)
		require.NoError(t, err)

		assert.NotEmpty(t, r.Header["Authorization"])
		assert.Equal(t, "Basic bWU6MTIzNDU=", r.Header["Authorization"][0])

		actual <- body
	}))
	t.Cleanup(srv.Close)

	requestConfig := fmt.Sprintf(`{
		"url": "%s",
		"method": "POST",
		"auth": {
			"type": "basic_auth",
			"config": {
				"user":     "me",
				"password": "12345"
			}
		},
		"body": "file://./stub/request.config.mailer.jsonnet"
	}`, srv.URL)

	_, reg := internal.NewFastRegistryWithMocks(t, configx.WithValues(map[string]any{
		config.ViperKeyCourierDeliveryStrategy:  "http",
		config.ViperKeyCourierHTTPRequestConfig: requestConfig,
		config.ViperKeyCourierSMTPURL:           "http://foo.url",
	}))

	courier, err := reg.Courier(t.Context())
	require.NoError(t, err)

	for _, message := range expectedEmail {
		id, err := courier.QueueEmail(t.Context(), email.NewTestStub(message))
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, id)
	}

	require.NoError(t, courier.DispatchQueue(t.Context()))
	close(actual)

	require.Len(t, actual, len(expectedEmail))

	i := 0
	for message := range actual {
		expected := expectedEmail[i]

		assert.Equal(t, expected.To, message.To)
		assert.Equal(t, expected.Body, message.Body)
		assert.Equal(t, expected.HTMLBody, message.HTMLBody)
		assert.Equal(t, expected.Subject, message.Subject)

		i++
	}
}
