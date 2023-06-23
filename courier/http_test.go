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
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
	"github.com/ory/x/resilience"
)

func TestQueueHTTPEmail(t *testing.T) {
	ctx := context.Background()

	type sendEmailRequestBody struct {
		IdentityID       string
		IdentityEmail    string
		Recipient        string
		TemplateType     string
		To               string
		RecoveryCode     string
		RecoveryURL      string
		VerificationURL  string
		VerificationCode string
		Body             string
		Subject          string
	}

	expectedEmail := []*email.TestStubModel{
		{
			To:      "test-2@test.com",
			Subject: "test-mailer-subject-1",
			Body:    "test-mailer-body-1",
		},
		{
			To:      "test-2@test.com",
			Subject: "test-mailer-subject-2",
			Body:    "test-mailer-body-2",
		},
	}

	actual := make([]sendEmailRequestBody, 0, 2)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rb, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var body sendEmailRequestBody

		err = json.Unmarshal(rb, &body)
		require.NoError(t, err)

		assert.NotEmpty(t, r.Header["Authorization"])
		assert.Equal(t, "Basic bWU6MTIzNDU=", r.Header["Authorization"][0])

		actual = append(actual, body)
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
		}
	}`, srv.URL)

	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeyCourierDeliveryStrategy, "http")
	conf.MustSet(ctx, config.ViperKeyCourierHTTPRequestConfig, requestConfig)
	conf.MustSet(ctx, config.ViperKeyCourierSMTPURL, "http://foo.url")
	reg.Logger().Level = logrus.TraceLevel

	courier, err := reg.Courier(ctx)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(ctx)
	defer t.Cleanup(cancel)

	for _, message := range expectedEmail {
		id, err := courier.QueueEmail(ctx, email.NewTestStub(reg, message))
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, id)
	}

	go func() {
		require.NoError(t, courier.Work(ctx))
	}()

	require.NoError(t, resilience.Retry(reg.Logger(), time.Millisecond*250, time.Second*10, func() error {
		if len(actual) == len(expectedEmail) {
			return nil
		}
		return errors.New("capacity not reached")
	}))

	for i, message := range actual {
		expected := email.NewTestStub(reg, expectedEmail[i])

		assert.Equal(t, x.Must(expected.EmailRecipient()), message.To)
		assert.Equal(t, x.Must(expected.EmailBody(ctx)), message.Body)
		assert.Equal(t, x.Must(expected.EmailSubject(ctx)), message.Subject)
	}
}
