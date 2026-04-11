package create_payment_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/money"
	app "github.com/soat13/payment/internal"
	"github.com/soat13/payment/internal/domain"
	messagingHandler "github.com/soat13/payment/internal/infra/in/messaging/create_payment"
	"github.com/soat13/payment/tests/integration"
	"github.com/stretchr/testify/require"
)

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
		message := iHaveAValidPaymentRequestMessage(t, externalID, amount)

		// When
		iReceiveTheSameMessageTwice(t, ctx, setup, message)

		// Then
		thePaymentShouldExistOnlyOnce(t, ctx, setup, externalID)
		thePaymentShouldBeCreatedAsPending(t, ctx, setup, externalID, amount)
	})
}

func iHaveAValidPaymentRequestMessage(t *testing.T, id uuid.UUID, amount money.Money) messagingHandler.Payload {
	t.Helper()

	return messagingHandler.Payload{
		ID:     id,
		Amount: amount,
	}
}

func iReceiveTheSameMessageTwice(t *testing.T, ctx context.Context, setup *integration.Setup, payloadMessage messagingHandler.Payload) {
	t.Helper()

	iReceiveAMessage(t, ctx, setup, payloadMessage)
	iReceiveAMessage(t, ctx, setup, payloadMessage)
}

func iReceiveAMessage(t *testing.T, ctx context.Context, setup *integration.Setup, payloadMessage messagingHandler.Payload) {
	t.Helper()

	encodedMessage, err := json.Marshal(payloadMessage)
	require.NoError(t, err)

	err = setup.Application.Broker.Publish(ctx, app.PaymentRequestQueue, encodedMessage)
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
