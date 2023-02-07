// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/tidwall/gjson"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/x"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/sqlcon"
)

type PersisterWrapper interface {
	GetConnection(ctx context.Context) *pop.Connection
	NetworkID(ctx context.Context) uuid.UUID
	courier.Persister
}

type NetworkWrapper func(t *testing.T, ctx context.Context) (uuid.UUID, PersisterWrapper)

func TestPersister(ctx context.Context, newNetworkUnlessExisting NetworkWrapper, newNetwork NetworkWrapper) func(t *testing.T) {
	return func(t *testing.T) {
		nid, p := newNetworkUnlessExisting(t, ctx)

		t.Run("case=no messages in queue", func(t *testing.T) {
			m, err := p.NextMessages(ctx, 10)
			require.ErrorIs(t, err, courier.ErrQueueEmpty)
			assert.Len(t, m, 0)

			_, err = p.LatestQueuedMessage(ctx)
			require.ErrorIs(t, err, courier.ErrQueueEmpty)
		})

		messages := make([]courier.Message, 5)
		t.Run("case=add messages to the queue", func(t *testing.T) {
			for k := range messages {
				require.NoError(t, faker.FakeData(&messages[k]))
				require.NoError(t, p.AddMessage(ctx, &messages[k]))
				time.Sleep(time.Second) // wait a bit so that the timestamp ordering works in MySQL.
			}
		})

		t.Run("case=latest message in queue", func(t *testing.T) {
			expected, err := p.LatestQueuedMessage(ctx)
			require.NoError(t, err)

			actual := messages[len(messages)-1]
			assert.Equal(t, expected.ID, actual.ID)
			assert.Equal(t, expected.Subject, actual.Subject)
		})

		t.Run("case=pull messages from the queue", func(t *testing.T) {
			for k, expected := range messages {
				expected.Status = courier.MessageStatusProcessing
				t.Run(fmt.Sprintf("message=%d", k), func(t *testing.T) {
					messages, err := p.NextMessages(ctx, 1)
					require.NoError(t, err)
					require.Len(t, messages, 1)

					actual := messages[0]
					assert.Equal(t, expected.ID, actual.ID)
					assert.Equal(t, expected.Subject, actual.Subject)
					assert.Equal(t, expected.Body, actual.Body)
					assert.Equal(t, expected.Status, actual.Status)
					assert.Equal(t, expected.Type, actual.Type)
					assert.Equal(t, expected.Recipient, actual.Recipient)
				})
			}

			_, err := p.NextMessages(ctx, 10)
			require.ErrorIs(t, err, courier.ErrQueueEmpty)
		})

		t.Run("case=setting message status", func(t *testing.T) {
			require.NoError(t, p.SetMessageStatus(ctx, messages[0].ID, courier.MessageStatusQueued))
			ms, err := p.NextMessages(ctx, 1)
			require.NoError(t, err)
			require.Len(t, ms, 1)
			assert.Equal(t, messages[0].ID, ms[0].ID)

			require.NoError(t, p.SetMessageStatus(ctx, messages[0].ID, courier.MessageStatusSent))
			_, err = p.NextMessages(ctx, 1)
			require.ErrorIs(t, err, courier.ErrQueueEmpty)

			require.NoError(t, p.SetMessageStatus(ctx, messages[0].ID, courier.MessageStatusAbandoned))
			_, err = p.NextMessages(ctx, 1)
			require.ErrorIs(t, err, courier.ErrQueueEmpty)
		})

		t.Run("case=incrementing send count", func(t *testing.T) {
			originalSendCount := messages[0].SendCount
			require.NoError(t, p.SetMessageStatus(ctx, messages[0].ID, courier.MessageStatusQueued))

			require.NoError(t, p.IncrementMessageSendCount(ctx, messages[0].ID))
			ms, err := p.NextMessages(ctx, 1)
			require.NoError(t, err)
			require.Len(t, ms, 1)
			assert.Equal(t, messages[0].ID, ms[0].ID)
			assert.Equal(t, originalSendCount+1, ms[0].SendCount)
		})

		t.Run("case=list messages", func(t *testing.T) {
			status := courier.MessageStatusProcessing
			filter := courier.ListCourierMessagesParameters{
				Status: &status,
			}
			ms, total, _, err := p.ListMessages(ctx, filter, []keysetpagination.Option{})

			require.NoError(t, err)
			assert.Len(t, ms, len(messages))
			assert.Equal(t, int64(len(messages)), total)
			assert.Equal(t, messages[len(messages)-1].ID, ms[0].ID)

			t.Run("on another network", func(t *testing.T) {
				nid1, p1 := newNetwork(t, ctx)
				ms, tc, _, err := p1.ListMessages(ctx, filter, []keysetpagination.Option{})

				require.NoError(t, err)
				require.Len(t, ms, 0)
				require.Equal(t, int64(0), tc)

				// Due to a bug in the pagination query definition, it was possible to retrieve messages from another `network`
				// using the pagination query. That required that 2 message's `created_at` timestamps were equal, to trigger
				// the `OR` clause of the paginated query.
				// This part of the tests "simulates" this behavior, by forcing the same timestamps on multiple messages across
				// different networks.
				nid2, p2 := newNetwork(t, ctx)
				const timeFormat = "2006-01-02 15:04:05.99999"
				msg1 := courier.Message{
					ID:     uuid.FromStringOrNil("10000000-0000-0000-0000-000000000000"),
					NID:    nid1,
					Status: courier.MessageStatusProcessing,
				}
				err = p1.GetConnection(ctx).Create(&msg1)
				require.NoError(t, err)

				msg2 := courier.Message{
					ID:     uuid.FromStringOrNil("20000000-0000-0000-0000-000000000000"),
					NID:    nid1,
					Status: courier.MessageStatusProcessing,
				}
				err = p1.GetConnection(ctx).Create(&msg2)
				require.NoError(t, err)
				msg3 := courier.Message{
					ID:     uuid.FromStringOrNil("30000000-0000-0000-0000-000000000000"),
					NID:    nid2,
					Status: courier.MessageStatusProcessing,
				}
				err = p2.GetConnection(ctx).Create(&msg3)
				require.NoError(t, err)
				now := time.Now().UTC().Truncate(time.Second).Format(timeFormat)

				// Set all `created_at` timestamps to the same value to force the `OR` clause of the paginated query.
				// `created_at` is set by "pop" and does not allow a manual override, apart from using `pop.SetNowFunc`, but that also influences the other tests in this
				// suite, as it just overrides a global function.
				require.NoError(t, p1.GetConnection(ctx).RawQuery("UPDATE courier_messages SET created_at = ? WHERE id = ? AND nid = ?", now, msg1.ID, nid1).Exec())
				// get the "updated" message from the
				require.NoError(t, p1.GetConnection(ctx).Where("id = ? AND nid = ?", msg1.ID, msg1.NID).First(&msg1))
				require.NoError(t, p1.GetConnection(ctx).RawQuery("UPDATE courier_messages SET created_at = ? WHERE id = ? AND nid = ?", now, msg2.ID, nid1).Exec())
				require.NoError(t, p2.GetConnection(ctx).RawQuery("UPDATE courier_messages SET created_at = ? WHERE id = ? AND nid = ?", now, msg3.ID, nid2).Exec())

				// Use the updated first message's PageToken as the basis for the paginated request.
				ms, _, _, err = p1.ListMessages(ctx, filter, []keysetpagination.Option{keysetpagination.WithToken(msg1.PageToken())})
				require.NoError(t, err)

				// The response should just contain the "next" message from network1, and not the message from network2
				require.Len(t, ms, 1)
				assert.Equal(t, ms[0].ID, msg2.ID)
			})
		})

		t.Run("case=network", func(t *testing.T) {

			t.Run("generates id on creation", func(t *testing.T) {
				expected := courier.Message{ID: uuid.Nil}
				require.NoError(t, p.AddMessage(ctx, &expected))

				assert.NotEqual(t, uuid.Nil, expected.ID)
				assert.EqualValues(t, nid, expected.NID)
				assert.EqualValues(t, nid, p.NetworkID(ctx))

				actual, err := p.LatestQueuedMessage(ctx)
				require.NoError(t, err)
				assert.EqualValues(t, expected.ID, actual.ID)
				assert.EqualValues(t, nid, actual.NID)

				actuals, err := p.NextMessages(ctx, 255)
				require.NoError(t, err)

				actual = &actuals[0]
				assert.EqualValues(t, expected.ID, actual.ID)
				assert.EqualValues(t, nid, actual.NID)
			})

			id := x.NewUUID()

			t.Run("persists id on creation", func(t *testing.T) {
				expected := courier.Message{ID: id}
				require.NoError(t, p.AddMessage(ctx, &expected))

				assert.EqualValues(t, id, expected.ID)
				assert.EqualValues(t, nid, expected.NID)
				assert.EqualValues(t, nid, p.NetworkID(ctx))

				actual, err := p.LatestQueuedMessage(ctx)
				require.NoError(t, err)
				assert.EqualValues(t, id, actual.ID)
				assert.EqualValues(t, nid, actual.NID)

				actuals, err := p.NextMessages(ctx, 255)
				require.NoError(t, err)

				actual = &actuals[0]
				assert.EqualValues(t, id, actual.ID)
				assert.EqualValues(t, nid, actual.NID)
			})

			t.Run("can not get on another network", func(t *testing.T) {
				_, p := newNetwork(t, ctx)

				_, err := p.LatestQueuedMessage(ctx)
				require.ErrorIs(t, err, courier.ErrQueueEmpty)

				_, err = p.NextMessages(ctx, 255)
				require.ErrorIs(t, err, courier.ErrQueueEmpty)
			})

			t.Run("can not update on another network", func(t *testing.T) {
				_, p := newNetwork(t, ctx)
				err := p.SetMessageStatus(ctx, id, courier.MessageStatusProcessing)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})

		})

		t.Run("case=FetchMessage", func(t *testing.T) {
			msgID := messages[0].ID

			message, err := p.FetchMessage(ctx, msgID)
			require.NoError(t, err)
			require.Equal(t, msgID, message.ID)

			t.Run("can not get on another network", func(t *testing.T) {
				_, p := newNetwork(t, ctx)

				_, err := p.FetchMessage(ctx, msgID)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})

		})

		t.Run("case=RecordDispatch", func(t *testing.T) {
			msgID := messages[0].ID

			err := p.RecordDispatch(ctx, msgID, courier.CourierMessageDispatchStatusFailed, errors.New("testerror"))
			require.NoError(t, err)

			message, err := p.FetchMessage(ctx, msgID)
			require.NoError(t, err)

			require.Len(t, message.Dispatches, 1)
			assert.Equal(t, "testerror", gjson.GetBytes(message.Dispatches[0].Error, "message").String())

			t.Run("can not get on another network", func(t *testing.T) {
				_, p := newNetwork(t, ctx)

				_, err := p.FetchMessage(ctx, msgID)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})
		})
	}
}
