package courier_test

import (
	"testing"

	"github.com/ory/kratos/courier"
	"github.com/stretchr/testify/require"
)

func TestMessageStatusValidity(t *testing.T) {
	invalid := courier.MessageStatus(0)
	require.Error(t, invalid.IsValid(), "IsValid() should return an error when message status is invalid")
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
		require.Error(t, err)
		require.Error(t, result.IsValid())
	})
}

func TestMessageTypeValidity(t *testing.T) {
	invalid := courier.MessageType(0)
	require.Error(t, invalid.IsValid(), "IsValid() should return an error when message type is invalid")
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
		require.Error(t, err)
		require.Error(t, result.IsValid())
	})
}
