package create_payment_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/payment/internal/application"
	"github.com/soat13/payment/internal/domain"
	messagingHandler "github.com/soat13/payment/internal/infra/in/messaging/create_payment"
	"github.com/soat13/payment/tests/integration"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type (
	event struct {
		messagingHandler.Payload
	}
)

func (e event) Name() string {
	return application.PaymentRequestQueue
}

func TestPaymentRequestFlow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	setup := integration.NewIntegrationSetup(t)

	setup.Application.Start(ctx)
	defer setup.Application.Stop()

	t.Run("should create a pending payment when a valid payment request message is received", func(t *testing.T) {
		externalID := uuid.New()
		amount, err := money.New(12345)
		require.NoError(t, err)

		// Given
		thePublisherExpectsAStatusChangedEvent(t, setup, externalID)
		message := iHaveAValidPaymentRequestMessage(t, externalID, amount)

		// When
		iReceiveAMessage(t, ctx, setup, message)

		// Then
		thePaymentShouldBeCreatedAsPending(t, ctx, setup, externalID, amount)
	})

	t.Run("should not create a duplicate payment when the same external_id is received twice", func(t *testing.T) {
		externalID := uuid.New()
		amount, err := money.New(12345)
		require.NoError(t, err)

		// Given
		thePublisherExpectsAStatusChangedEvent(t, setup, externalID)
		message := iHaveAValidPaymentRequestMessage(t, externalID, amount)

		// When
		iReceiveTheSameMessageTwice(t, ctx, setup, message)

		// Then
		thePaymentShouldExistOnlyOnce(t, ctx, setup, externalID)
		thePaymentShouldBeCreatedAsPending(t, ctx, setup, externalID, amount)
	})
}

func iHaveAValidPaymentRequestMessage(t *testing.T, id uuid.UUID, amount money.Money) event {
	t.Helper()

	return event{
		Payload: messagingHandler.Payload{
			ID:     id,
			Amount: amount,
		},
	}
}

func iReceiveTheSameMessageTwice(t *testing.T, ctx context.Context, setup *integration.Setup, message event) {
	t.Helper()

	iReceiveAMessage(t, ctx, setup, message)
	iReceiveAMessage(t, ctx, setup, message)
}

func iReceiveAMessage(t *testing.T, ctx context.Context, setup *integration.Setup, message event) {
	t.Helper()

	err := setup.Application.QueuePublisher.Send(ctx, message)
	require.NoError(t, err)
}

func thePaymentShouldExistOnlyOnce(t *testing.T, ctx context.Context, setup *integration.Setup, externalID uuid.UUID) {
	t.Helper()

	total, err := setup.Application.Repository.CountByExternalID(ctx, externalID)
	require.NoError(t, err)
	require.Equal(t, 1, total)
}

func thePaymentShouldBeCreatedAsPending(t *testing.T, ctx context.Context, setup *integration.Setup, externalID uuid.UUID, amount money.Money) {
	t.Helper()

	payment, err := setup.Application.Repository.GetByExternalID(ctx, externalID)
	require.NoError(t, err)
	require.NotNil(t, payment)
	require.Equal(t, externalID, payment.ExternalID)
	require.Equal(t, amount.Cents, payment.Amount.Cents)
	require.Equal(t, domain.StatusPending, payment.Status)
}

func thePublisherExpectsAStatusChangedEvent(t *testing.T, setup *integration.Setup, externalID uuid.UUID) {
	t.Helper()

	setup.TopicPublisher.
		EXPECT().
		Publish(gomock.Any(), gomock.AssignableToTypeOf(domain.StatusChangedEvent{})).
		DoAndReturn(func(_ context.Context, event domain.Event) error {
			statusChanged, ok := event.(domain.StatusChangedEvent)
			require.True(t, ok)

			require.Equal(t, "payment-status-changed", statusChanged.Name())
			require.Equal(t, externalID, statusChanged.ExternalID)
			require.Equal(t, domain.StatusPending, statusChanged.Status)
			require.NotEqual(t, uuid.Nil, statusChanged.ID)

			return nil
		}).
		Times(1)
}
