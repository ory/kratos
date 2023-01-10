// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/kratos/courier"
)

func TestMessageStatusValidity(t *testing.T) {
	invalid := courier.MessageStatus(0)
	require.ErrorIs(t, invalid.IsValid(), herodot.ErrBadRequest, "IsValid() should return an error when message status is invalid")
}

func TestToMessageStatus(t *testing.T) {
	t.Run("case=should return corresponding MessageStatus for given str", func(t *testing.T) {
		for str, exp := range map[string]courier.MessageStatus{
			"queued":     courier.MessageStatusQueued,
			"sent":       courier.MessageStatusSent,
			"processing": courier.MessageStatusProcessing,
			"abandoned":  courier.MessageStatusAbandoned,
		} {
			result, err := courier.ToMessageStatus(str)
			require.NoError(t, err)
			require.Equal(t, exp, result)
		}
	})
	t.Run("case=should return error for invalid message status str", func(t *testing.T) {
		result, err := courier.ToMessageStatus("invalid")
		require.Error(t, err, herodot.ErrBadRequest)
		require.Error(t, result.IsValid(), herodot.ErrBadRequest)
	})
}

func TestMessageTypeValidity(t *testing.T) {
	invalid := courier.MessageType(0)
	require.ErrorIs(t, invalid.IsValid(), herodot.ErrBadRequest, "IsValid() should return an error when message type is invalid")
}

func TestToMessageType(t *testing.T) {
	t.Run("case=should return corresponding MessageType for given str", func(t *testing.T) {
		for str, exp := range map[string]courier.MessageType{
			"email": courier.MessageTypeEmail,
			"phone": courier.MessageTypePhone,
		} {
			result, err := courier.ToMessageType(str)
			require.NoError(t, err)
			require.Equal(t, exp, result)
		}
	})
	t.Run("case=should return error for invalid message type str", func(t *testing.T) {
		result, err := courier.ToMessageType("invalid")
		require.ErrorIs(t, err, herodot.ErrBadRequest)
		require.ErrorIs(t, result.IsValid(), herodot.ErrBadRequest)
	})
}
