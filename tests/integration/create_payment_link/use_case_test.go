package create_link_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/ms-payment/internal/application/ports/out"
	"github.com/soat13/ms-payment/internal/domain"
	messagingHandler "github.com/soat13/ms-payment/internal/infra/in/messaging/create_link"
	"github.com/soat13/ms-payment/tests/integration"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type (
	queueMessage struct {
		messagingHandler.Payload
	}
)

func (e queueMessage) Name() string {
	return "payment-link-request"
}

func TestCreatePaymentLinkFlow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	setup := integration.NewIntegrationSetup(t)

	setup.Application.Start(ctx, false)
	defer setup.Application.Stop()

	t.Run("should create payment link when payment status changed event is pending", func(t *testing.T) {
		externalID := uuid.New()
		amount, err := money.New(12345)
		require.NoError(t, err)

		// Given
		payment := iHaveAPersistedPendingPayment(t, ctx, setup, externalID, amount)
		theServiceExpectsPaymentLinkCreation(t, setup, payment.ID)
		thePublisherExpectsAStatusChangedEventWithProcessingStatus(t, setup, payment.ID, externalID)

		message := iHaveAPendingMessage(payment.ID)

		// When
		iReceiveAMessage(t, ctx, setup, message)

		// Then
		thePaymentShouldBeProcessingWithLink(t, ctx, setup, payment.ID)
	})

	t.Run("should ignore message when payment status is not pending", func(t *testing.T) {
		externalID := uuid.New()
		amount, err := money.New(12345)
		require.NoError(t, err)

		// Given
		payment := iHaveAPersistedPendingPayment(t, ctx, setup, externalID, amount)
		message := iHaveAProcessingMessage(payment.ID)

		// When
		iReceiveAMessage(t, ctx, setup, message)

		// Then
		thePaymentShouldRemainPendingWithoutLink(t, ctx, setup, payment.ID)
	})
}

func iHaveAPendingMessage(id uuid.UUID) queueMessage {
	return createAValidMessage(id, domain.StatusPending)
}

func iHaveAProcessingMessage(id uuid.UUID) queueMessage {
	return createAValidMessage(id, domain.StatusProcessing)
}

func createAValidMessage(id uuid.UUID, status domain.Status) queueMessage {
	return queueMessage{
		Payload: messagingHandler.Payload{
			ID:     id,
			Status: status,
		},
	}
}

func iHaveAPersistedPendingPayment(
	t *testing.T,
	ctx context.Context,
	setup *integration.Setup,
	externalID uuid.UUID,
	amount money.Money,
) *domain.Payment {
	t.Helper()

	payment, err := domain.NewPendingPayment(externalID, amount, "Some description")
	require.NoError(t, err)

	err = setup.Container.Repository.Save(ctx, payment)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, payment.ID)

	return payment
}

func iReceiveAMessage(t *testing.T, ctx context.Context, setup *integration.Setup, message queueMessage) {
	t.Helper()

	err := setup.Container.QueueSender.Send(ctx, message)
	require.NoError(t, err)
}

func thePaymentShouldBeProcessingWithLink(
	t *testing.T,
	ctx context.Context,
	setup *integration.Setup,
	paymentID uuid.UUID,
) {
	t.Helper()

	payment, err := setup.Container.Repository.GetByID(ctx, paymentID)
	require.NoError(t, err)
	require.NotNil(t, payment)

	require.Equal(t, domain.StatusProcessing, payment.Status)
	require.NotNil(t, payment.Link)
	require.NotNil(t, payment.Provider)
	require.NotNil(t, payment.ProviderPaymentID)

	require.Equal(t, domain.MercadoPagoProviderName, *payment.Provider)
	require.Equal(t, "https://fake.mercadopago.com/checkout/123", string(*payment.Link))
	require.Equal(t, "provider-payment-id-123", string(*payment.ProviderPaymentID))
}

func thePaymentShouldRemainPendingWithoutLink(
	t *testing.T,
	ctx context.Context,
	setup *integration.Setup,
	paymentID uuid.UUID,
) {
	t.Helper()

	payment, err := setup.Container.Repository.GetByID(ctx, paymentID)
	require.NoError(t, err)
	require.NotNil(t, payment)

	require.Equal(t, domain.StatusPending, payment.Status)
	require.Nil(t, payment.Link)
	require.Nil(t, payment.Provider)
	require.Nil(t, payment.ProviderPaymentID)
}

func theServiceExpectsPaymentLinkCreation(t *testing.T, setup *integration.Setup, paymentID uuid.UUID) {
	t.Helper()

	setup.MockPaymentProvider.
		EXPECT().
		CreateLink(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input out.CreateLinkInput) (*out.CreateLinkOutput, error) {
			require.Equal(t, paymentID, input.PaymentID)

			return &out.CreateLinkOutput{
				Link:       "https://fake.mercadopago.com/checkout/123",
				Provider:   domain.MercadoPagoProviderName,
				ProviderID: "provider-payment-id-123",
			}, nil
		}).
		Times(1)
}

func thePublisherExpectsAStatusChangedEventWithProcessingStatus(
	t *testing.T,
	setup *integration.Setup,
	paymentID uuid.UUID,
	externalID uuid.UUID,
) {
	t.Helper()

	setup.MockTopicPublisher.
		EXPECT().
		Publish(gomock.Any(), gomock.AssignableToTypeOf(domain.StatusChangedEvent{})).
		DoAndReturn(func(_ context.Context, event domain.Event) error {
			statusChanged, ok := event.(domain.StatusChangedEvent)
			require.True(t, ok)

			require.Equal(t, "payment-status-changed.fifo", statusChanged.Name())
			require.Equal(t, paymentID, statusChanged.ID)
			require.Equal(t, externalID, statusChanged.ExternalID)
			require.Equal(t, domain.StatusProcessing, statusChanged.Status)
			require.Equal(t, "https://fake.mercadopago.com/checkout/123", *statusChanged.PaymentURL)

			return nil
		}).
		Times(1)
}
