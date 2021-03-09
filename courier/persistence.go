package courier

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofrs/uuid"
)

var ErrQueueEmpty = errors.New("queue is empty")

type (
	Persister interface {
		AddMessage(context.Context, *Message) error

		NextMessages(context.Context, uint8) ([]Message, error)

		SetMessageStatus(context.Context, uuid.UUID, MessageStatus) error

		LatestQueuedMessage(ctx context.Context) (*Message, error)
	}

	PersistenceProvider interface {
		CourierPersister() Persister
	}
)

func TestPersister(ctx context.Context, p Persister) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("case=no messages in queue", func(t *testing.T) {
			m, err := p.NextMessages(ctx, 10)
			require.EqualError(t, err, ErrQueueEmpty.Error())
			assert.Len(t, m, 0)

			_, err = p.LatestQueuedMessage(ctx)
			require.EqualError(t, err, ErrQueueEmpty.Error())
		})

		messages := make([]Message, 5)
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
				expected.Status = MessageStatusProcessing
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
			require.EqualError(t, err, ErrQueueEmpty.Error())
		})

		t.Run("case=setting message status", func(t *testing.T) {
			require.NoError(t, p.SetMessageStatus(ctx, messages[0].ID, MessageStatusQueued))
			ms, err := p.NextMessages(ctx, 1)
			require.NoError(t, err)
			require.Len(t, ms, 1)
			assert.Equal(t, messages[0].ID, ms[0].ID)

			require.NoError(t, p.SetMessageStatus(ctx, messages[0].ID, MessageStatusSent))
			_, err = p.NextMessages(ctx, 1)
			require.EqualError(t, err, ErrQueueEmpty.Error())
		})
	}
}
