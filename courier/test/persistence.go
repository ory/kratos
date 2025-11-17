// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/x"
	"github.com/ory/pop/v6"
	keysetpagination "github.com/ory/x/pagination/keysetpagination_v2"
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
			t.Cleanup(func() { pop.SetNowFunc(func() time.Time { return time.Now().Round(time.Second) }) })
			now := time.Now()
			for k := range messages {
				// We need to fake the time func to control the created_at column, which is the
				// sort key for the messages.
				pop.SetNowFunc(func() time.Time { return now.Add(time.Duration(k) * time.Hour).Round(time.Second) })
				require.NoError(t, faker.FakeData(&messages[k]))
				require.NoError(t, p.AddMessage(ctx, &messages[k]))
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
			// List by status.
			{
				status := courier.MessageStatusProcessing
				filter := courier.ListCourierMessagesParameters{
					Status: &status,
				}
				ms, _, err := p.ListMessages(ctx, filter, []keysetpagination.Option{})

				require.NoError(t, err)
				require.Len(t, ms, len(messages), messages)
				// Check that the 'filter by status' works.
				for _, m := range ms {
					require.Equal(t, status, m.Status)
				}

				// Check that the 'order by created_at desc' works.
				require.True(t, slices.IsSortedFunc(ms, func(a, b courier.Message) int { return b.CreatedAt.Compare(a.CreatedAt) }))
			}
			// Query fewer items than the total, multiple times.
			{
				filter := courier.ListCourierMessagesParameters{}
				maxSize := 2
				ms1, pagination, err := p.ListMessages(ctx, filter, []keysetpagination.Option{
					keysetpagination.WithSize(2),
				})
				require.NoError(t, err)
				require.NotNil(t, pagination)
				require.False(t, pagination.IsLast())
				require.Len(t, ms1, maxSize)

				// Check that the 'order by created_at desc' works.
				require.True(t, slices.IsSortedFunc(ms1, func(a, b courier.Message) int { return b.CreatedAt.Compare(a.CreatedAt) }))

				// Second call.
				// Marshal -> unmarshal the pagination token to be more realistic.
				encrypted := pagination.PageToken().Encrypt(nil)
				unmarshalled, err := keysetpagination.ParsePageToken(nil, encrypted)
				require.NoError(t, err)

				ms2, pagination, err := p.ListMessages(ctx, filter,
					[]keysetpagination.Option{
						keysetpagination.WithSize(2),
						keysetpagination.WithToken(unmarshalled),
					})

				require.NoError(t, err)
				require.NotNil(t, pagination)
				require.False(t, pagination.IsLast())
				require.Len(t, ms2, maxSize)
				// Check that the 'order by created_at desc' works.
				require.True(t, slices.IsSortedFunc(ms2, func(a, b courier.Message) int { return b.CreatedAt.Compare(a.CreatedAt) }))

				// Check that the second call returned different elements.
				require.NotEqual(t, ms1[0].ID, ms2[0].ID)
				require.NotEqual(t, ms1[1].ID, ms2[1].ID)
				allElements := append(ms1, ms2...)
				require.True(t, slices.IsSortedFunc(allElements, func(a, b courier.Message) int { return b.CreatedAt.Compare(a.CreatedAt) }))

				// Last call
				ms3, pagination, err := p.ListMessages(ctx, filter, pagination.ToOptions())
				require.NoError(t, err)
				require.NotNil(t, pagination)
				require.True(t, pagination.IsLast())
				require.Len(t, ms3, 1)
				// Check that the 'order by created_at desc' works.
				require.True(t, slices.IsSortedFunc(ms3, func(a, b courier.Message) int { return b.CreatedAt.Compare(a.CreatedAt) }))

				// Check that the third call returned different elements.
				require.NotEqual(t, ms2[0].ID, ms3[0].ID)
				allElements = append(ms1, ms2...)
				allElements = append(allElements, ms3...)
				require.True(t, slices.IsSortedFunc(allElements, func(a, b courier.Message) int { return b.CreatedAt.Compare(a.CreatedAt) }))
			}

			t.Run("on another network", func(t *testing.T) {
				nid1, p1 := newNetwork(t, ctx)
				status := courier.MessageStatusProcessing
				filter := courier.ListCourierMessagesParameters{
					Status: &status,
				}
				ms, _, err := p1.ListMessages(ctx, filter, []keysetpagination.Option{})

				require.NoError(t, err)
				require.Len(t, ms, 0)

				// Due to a bug in the pagination query definition, it was possible to retrieve messages from another `network`
				// using the pagination query. That required that 2 message's `created_at` timestamps were equal, to trigger
				// the `OR` clause of the paginated query.
				// This part of the tests "simulates" this behavior, by forcing the same timestamps on multiple messages across
				// different networks.
				nid2, p2 := newNetwork(t, ctx)
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
				now := time.Now().UTC().Truncate(time.Second)

				// Set all `created_at` timestamps to the same value to force the `OR` clause of the paginated query.
				// `created_at` is set by "pop" and does not allow a manual override, apart from using `pop.SetNowFunc`, but that also influences the other tests in this
				// suite, as it just overrides a global function.
				require.NoError(t, p1.GetConnection(ctx).RawQuery("UPDATE courier_messages SET created_at = ? WHERE id = ? AND nid = ?", now, msg1.ID, nid1).Exec())
				// get the "updated" message from the
				require.NoError(t, p1.GetConnection(ctx).Where("id = ? AND nid = ?", msg1.ID, msg1.NID).First(&msg1))
				require.NoError(t, p1.GetConnection(ctx).RawQuery("UPDATE courier_messages SET created_at = ? WHERE id = ? AND nid = ?", now, msg2.ID, nid1).Exec())
				require.NoError(t, p2.GetConnection(ctx).RawQuery("UPDATE courier_messages SET created_at = ? WHERE id = ? AND nid = ?", now, msg3.ID, nid2).Exec())

				// Use the updated first message's PageToken as the basis for the paginated request.
				ms, _, err = p1.ListMessages(ctx, filter, []keysetpagination.Option{keysetpagination.WithToken(msg1.PageToken())})
				require.NoError(t, err)

				// The response should just contain messages from network1, and not from network2.
				require.Len(t, ms, 1)
				require.Equal(t, ms[0].NID, nid1)
				require.Equal(t, ms[0].ID, msg2.ID)
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
