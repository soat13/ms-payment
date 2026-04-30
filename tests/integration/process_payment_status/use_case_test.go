package process_payment_status_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/ms-payment/internal/domain"
	"github.com/soat13/ms-payment/internal/infra/out/providers/mercado_pago"
	"github.com/soat13/ms-payment/tests/integration"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestProcessPaymentStatusFlow(t *testing.T) {
	ctx := context.Background()
	setup := integration.NewIntegrationSetup(t)

	setup.Application.Start(ctx, true)
	defer setup.Application.Stop()

	t.Run("should process payment status when mercado pago merchant order webhook is received", func(t *testing.T) {
		externalID := uuid.New()
		amount := createMoney(t, 12345)

		// Given
		payment := iHaveAProcessingPayment(t, ctx, setup, externalID, amount)
		thePaymentProviderFindsMerchantOrderPaymentAsSucceeded(t, setup, payment.ID)
		thePublisherExpectsAStatusChangedEventWithSucceededStatus(t, setup, payment.ID, externalID)

		// When
		response := iReceiveAMercadoPagoMerchantOrderWebhook(t, setup)

		// Then
		theResponseShouldBeOK(t, response)
		thePaymentShouldBeSucceeded(t, ctx, setup, payment.ID)
	})

	t.Run("should ignore unsupported webhook payload", func(t *testing.T) {
		// When
		response := iReceiveAnUnsupportedMercadoPagoWebhook(t, setup)

		// Then
		theResponseShouldBeNoContent(t, response)
	})

	t.Run("should return bad request when merchant order resource is invalid", func(t *testing.T) {
		// When
		response := iReceiveAnInvalidMercadoPagoMerchantOrderWebhook(t, setup)

		// Then
		theResponseShouldBeBadRequest(t, response)
	})
}

func createMoney(t *testing.T, cents int64) money.Money {
	t.Helper()

	amount, err := money.New(cents)
	require.NoError(t, err)

	return amount
}

func iHaveAProcessingPayment(
	t *testing.T,
	ctx context.Context,
	setup *integration.Setup,
	externalID uuid.UUID,
	amount money.Money,
) *domain.Payment {
	t.Helper()

	payment, err := domain.NewPendingPayment(externalID, amount, "Some description")
	require.NoError(t, err)

	err = payment.StartProcessing(
		"https://fake.mercadopago.com/checkout/123",
		domain.MercadoPagoProviderName,
		"provider-payment-id-123",
	)
	require.NoError(t, err)

	err = setup.Container.Repository.Save(ctx, payment)
	require.NoError(t, err)

	return payment
}

func iReceiveAMercadoPagoMerchantOrderWebhook(
	t *testing.T,
	setup *integration.Setup,
) *http.Response {
	t.Helper()

	return iReceiveAMercadoPagoWebhook(t, setup, `{
		"resource": "https://api.mercadolibre.com/merchant_orders/40165307899",
		"topic": "merchant_order"
	}`)
}

func iReceiveAnUnsupportedMercadoPagoWebhook(
	t *testing.T,
	setup *integration.Setup,
) *http.Response {
	t.Helper()

	return iReceiveAMercadoPagoWebhook(t, setup, `{
		"resource": "https://api.mercadolibre.com/payments/123",
		"topic": "payment"
	}`)
}

func iReceiveAnInvalidMercadoPagoMerchantOrderWebhook(
	t *testing.T,
	setup *integration.Setup,
) *http.Response {
	t.Helper()

	return iReceiveAMercadoPagoWebhook(t, setup, `{
		"resource": "https://api.mercadolibre.com/merchant_orders/invalid",
		"topic": "merchant_order"
	}`)
}

func iReceiveAMercadoPagoWebhook(
	t *testing.T,
	setup *integration.Setup,
	body string,
) *http.Response {
	t.Helper()

	request := httptest.NewRequest(
		fiber.MethodPost,
		"/webhooks/mercado-pago",
		bytes.NewBufferString(body),
	)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	response, err := setup.FiberApp.Test(request)
	require.NoError(t, err)

	return response
}

func thePaymentProviderFindsMerchantOrderPaymentAsSucceeded(
	t *testing.T,
	setup *integration.Setup,
	paymentID uuid.UUID,
) {
	t.Helper()

	setup.MockMercadoPagoClient.
		EXPECT().
		FindByMerchantID(gomock.Any(), 40165307899).
		DoAndReturn(func(_ context.Context, merchantOrderID int) (*mercado_pago.ProcessPaymentStatusResponse, error) {
			require.Equal(t, 40165307899, merchantOrderID)

			return &mercado_pago.ProcessPaymentStatusResponse{
				PaymentID: paymentID,
				Status:    domain.StatusSucceeded,
			}, nil
		}).
		Times(1)
}

func thePublisherExpectsAStatusChangedEventWithSucceededStatus(
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
			require.Equal(t, domain.StatusSucceeded, statusChanged.Status)

			return nil
		}).
		Times(1)
}

func thePaymentShouldBeSucceeded(
	t *testing.T,
	ctx context.Context,
	setup *integration.Setup,
	paymentID uuid.UUID,
) {
	t.Helper()

	payment, err := setup.Container.Repository.GetByID(ctx, paymentID)
	require.NoError(t, err)
	require.NotNil(t, payment)

	require.Equal(t, domain.StatusSucceeded, payment.Status)
}

func theResponseShouldBeOK(t *testing.T, response *http.Response) {
	t.Helper()

	require.Equal(t, fiber.StatusOK, response.StatusCode)
}

func theResponseShouldBeNoContent(t *testing.T, response *http.Response) {
	t.Helper()

	require.Equal(t, fiber.StatusNoContent, response.StatusCode)
}

func theResponseShouldBeBadRequest(t *testing.T, response *http.Response) {
	t.Helper()

	require.Equal(t, fiber.StatusBadRequest, response.StatusCode)
}
