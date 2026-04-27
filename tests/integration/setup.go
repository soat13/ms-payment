package integration

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/soat13/payment/internal/application/ports/out/mock"
	"github.com/soat13/payment/internal/infra"
	"github.com/soat13/payment/internal/infra/bootstrap"
	mockMercadoPagoClient "github.com/soat13/payment/internal/infra/out/providers/mercado_pago/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type Setup struct {
	Application           *infra.App
	Container             *bootstrap.Container
	MockTopicPublisher    *mock.MockTopicPublisher
	MockPaymentProvider   *mock.MockPaymentProvider
	MockMercadoPagoClient *mockMercadoPagoClient.MockClient
	FiberApp              *fiber.App
}

func NewIntegrationSetup(t *testing.T) *Setup {
	t.Helper()

	env := bootstrap.Envs{
		IsTest:        true,
		AwsRegion:     "us-east-1",
		AwsEndpoint:   "http://localhost:4566",
		AwsBaseSNSARN: "arn:aws:sns:us-east-1:000000000000",
		DynamodbGSI:   "gsi1",
		DynamodbTable: "payments_test",
		HttpPort:      "8181",
	}

	ctx := context.Background()
	container, err := bootstrap.NewContainer(ctx, &env)
	require.NoError(t, err)

	topicPublisher := mock.NewMockTopicPublisher(gomock.NewController(t))
	paymentProvider := mock.NewMockPaymentProvider(gomock.NewController(t))
	mercadoPagoClient := mockMercadoPagoClient.NewMockClient(gomock.NewController(t))

	container.TopicPublisher = topicPublisher
	container.PaymentProvider = paymentProvider
	container.MercadoPagoClient = mercadoPagoClient

	return &Setup{
		Container:             container,
		Application:           infra.NewApp(container),
		MockTopicPublisher:    topicPublisher,
		MockPaymentProvider:   paymentProvider,
		MockMercadoPagoClient: mercadoPagoClient,
		FiberApp:              container.FiberApp,
	}
}
