package messaging

import (
	"testing"

	"github.com/google/uuid"
	"github.com/soat13/ms-payment/internal/domain"
	"github.com/stretchr/testify/require"
)

type simpleEvent struct{}

func (s simpleEvent) Name() string {
	return "simple.event"
}

func TestExtractGroupID(t *testing.T) {
	t.Parallel()

	t.Run("should return group id for grouped event", func(t *testing.T) {
		t.Parallel()

		externalID := uuid.New()

		event := domain.StatusChangedEvent{
			ExternalID: externalID,
		}

		result := ExtractGroupID(event)

		require.NotNil(t, result)
		require.Equal(t, externalID.String(), *result)
	})

	t.Run("should return nil for non grouped event", func(t *testing.T) {
		t.Parallel()

		event := simpleEvent{}

		result := ExtractGroupID(event)

		require.Nil(t, result)
	})
}
