// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier"
	"github.com/ory/x/pagination/keysetpagination"
)

func CourierExpectMessage(ctx context.Context, t *testing.T, reg interface {
	courier.PersistenceProvider
}, recipient, subject string,
) *courier.Message {
	messages, total, _, err := reg.CourierPersister().ListMessages(ctx, courier.ListCourierMessagesParameters{
		Recipient: recipient,
	}, []keysetpagination.Option{})
	require.NoError(t, err)
	require.GreaterOrEqual(t, total, int64(1))

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt.After(messages[j].CreatedAt)
	})

	for _, m := range messages {
		if strings.EqualFold(m.Recipient, recipient) && strings.EqualFold(m.Subject, subject) {
			return &m
		}
	}

	require.Failf(t, "could not find courier messages with recipient %s and subject %s", recipient, subject)
	return nil
}

func CourierExpectLinkInMessage(t *testing.T, message *courier.Message, offset int) string {
	if offset == 0 {
		offset = 1
	}
	match := regexp.MustCompile(`(http[^\s]+)`).FindStringSubmatch(message.Body)
	require.Len(t, match, offset*2)

	return match[offset]
}

func CourierExpectCodeInMessage(t *testing.T, message *courier.Message, offset int) string {
	if offset == 0 {
		offset = 1
	}
	match := regexp.MustCompile(CodeRegex).FindStringSubmatch(message.Body)
	require.Len(t, match, offset*2)

	return match[offset]
}
