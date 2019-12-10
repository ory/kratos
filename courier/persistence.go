package courier

import (
	"context"
	"fmt"
	"testing"

	"github.com/bxcodec/faker"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/identity"
)

var ErrQueueEmpty = errors.New("queue is empty")

type (
	Persister interface {
		AddMessage(context.Context, *Message) error

		NextMessages(context.Context, uint8) ([]Message, error)

		SetMessageStatus(context.Context, uuid.UUID, MessageStatus) error
	}
	PersistenceProvider interface {
		CourierPersister() Persister
	}
)

func TestPersister(p interface {
	Persister
	identity.Pool
}) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("case=no messages in queue", func(t *testing.T) {
			m, err := p.NextMessages(context.Background(), 10)
			require.EqualError(t, err, ErrQueueEmpty.Error())
			assert.Len(t, m, 0)
		})

		messages := make([]Message, 5)
		t.Run("case=add messages to the queue", func(t *testing.T) {
			for k, m := range messages {
				require.NoError(t, faker.FakeData(&m))
				require.NoError(t, p.AddMessage(context.Background(), &m))
				messages[k] = m
			}
		})

		t.Run("case=pull messages from the queue", func(t *testing.T) {
			for k, expected := range messages {
				t.Run(fmt.Sprintf("message=%d", k), func(t *testing.T) {
					messages, err := p.NextMessages(context.Background(), 1)
					require.NoError(t, err)
					require.Len(t, messages, 1)

					actual := messages[0]
					assert.Equal(t, expected.ID, actual.ID)
					assert.Equal(t, expected.Subject, actual.Subject)
					assert.Equal(t, expected.Body, actual.Body)
					assert.Equal(t, expected.Status, actual.Status)
					assert.Equal(t, expected.Type, actual.Type)
					assert.Equal(t, expected.Recipient, actual.Recipient)

					require.NoError(t, p.SetMessageStatus(context.Background(), actual.ID, MessageStatusSent))
				})
			}

			ms, err := p.NextMessages(context.Background(), 10)
			require.NoError(t, err)
			assert.Len(t, ms, 0)
		})

		t.Run("case=setting message status", func(t *testing.T) {
			require.NoError(t, p.SetMessageStatus(context.Background(), messages[0].ID, MessageStatusQueued))
			ms, err := p.NextMessages(context.Background(), 1)
			require.NoError(t, err)
			require.Len(t, ms, 1)
			assert.Equal(t, messages[0].ID, ms[1].ID)

			require.NoError(t, p.SetMessageStatus(context.Background(), messages[0].ID, MessageStatusSent))
			ms, err = p.NextMessages(context.Background(), 1)
			require.EqualError(t, err, ErrQueueEmpty.Error())
			assert.Len(t, ms, 0)
		})
	}
}
