package testhelpers

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier"
)

func CourierExpectMessage(t *testing.T, reg interface {
	courier.PersistenceProvider
}, recipient, subject string) *courier.Message {
	message, err := reg.CourierPersister().LatestQueuedMessage(context.Background())
	require.NoError(t, err)

	assert.EqualValues(t, subject, strings.TrimSpace(message.Subject))
	assert.EqualValues(t, recipient, strings.TrimSpace(message.Recipient))

	return message
}

func CourierExpectLinkInMessage(t *testing.T, message *courier.Message, offset int) string {
	if offset == 0 {
		offset++
	}
	match := regexp.MustCompile(`(http[^\s]+)`).FindStringSubmatch(message.Body)
	require.Len(t, match, offset*2)

	return match[offset]
}
