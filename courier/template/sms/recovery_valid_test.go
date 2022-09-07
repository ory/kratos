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

func TestNewRecoveryOTPMessage(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)

	const (
		expectedPhone = "+12345678901"
		otp           = "012345"
	)

	tpl := sms.NewRecoveryOTPMessage(reg, &sms.RecoveryMessageModel{To: expectedPhone, Code: otp})

	expectedBody := fmt.Sprintf("Your recovery code is: %s.\nDon't share this code with anyone.\n", otp)

	actualBody, err := tpl.SMSBody(context.Background())
	require.NoError(t, err)
	assert.Equal(t, expectedBody, actualBody)

	actualPhone, err := tpl.PhoneNumber()
	require.NoError(t, err)
	assert.Equal(t, expectedPhone, actualPhone)
}
