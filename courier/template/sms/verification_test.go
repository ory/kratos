package sms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier/template/sms"
	"github.com/ory/kratos/internal"
)

func TestNewVerificationMessage(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)

	const (
		expectedPhone           = "+12345678901"
		expectedVerificationURL = "http://bar.foo"
	)

	tpl := sms.NewVerificationMessage(reg, &sms.VerificationMessageModel{To: expectedPhone, VerificationURL: expectedVerificationURL})

	expectedBody := fmt.Sprintf("Hi, please verify your account by clicking the following link:\n\n%s\n", expectedVerificationURL)

	actualBody, err := tpl.SMSBody(context.Background())
	require.NoError(t, err)
	assert.Equal(t, expectedBody, actualBody)

	actualPhone, err := tpl.PhoneNumber()
	require.NoError(t, err)
	assert.Equal(t, expectedPhone, actualPhone)
}
