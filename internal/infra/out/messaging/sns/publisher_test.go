package sns

import (
	"context"
	"errors"
	"testing"

	"github.com/soat13/oficina-utils/pkg/messaging"
	"github.com/stretchr/testify/require"
)

type fakeTopicPublisher struct {
	message messaging.TopicMessage
	err     error
}

func (f *fakeTopicPublisher) Publish(ctx context.Context, message messaging.TopicMessage) error {
	f.message = message
	return f.err
}

type fakeEvent struct {
	name string
}

func (f fakeEvent) Name() string {
	return f.name
}

func TestPublisherPublish(t *testing.T) {
	t.Parallel()

	t.Run("should publish message", func(t *testing.T) {
		t.Parallel()

		fakePublisher := &fakeTopicPublisher{}

		publisher := NewPublisher(fakePublisher)

		event := fakeEvent{name: "payment.created"}

		err := publisher.Publish(context.Background(), event)

		require.NoError(t, err)
		require.Equal(t, "payment.created", fakePublisher.message.EventName)
		require.Equal(t, event, fakePublisher.message.Payload)
	})

	t.Run("should return error when publisher fails", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("publish error")

		fakePublisher := &fakeTopicPublisher{
			err: expectedErr,
		}

		publisher := NewPublisher(fakePublisher)

		err := publisher.Publish(context.Background(), fakeEvent{name: "payment.created"})

		require.ErrorIs(t, err, expectedErr)
	})
}
