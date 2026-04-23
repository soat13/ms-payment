package integration

import (
	"context"
	"testing"

	"github.com/soat13/payment/internal/application/ports/out/mock"
	"github.com/soat13/payment/internal/infra/bootstrap"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type Setup struct {
	Application         *bootstrap.App
	Container           *bootstrap.Container
	MockTopicPublisher  *mock.MockTopicPublisher
	MockPaymentProvider *mock.MockPaymentProvider
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
	}

	ctx := context.Background()
	container, err := bootstrap.NewContainer(ctx, &env)
	require.NoError(t, err)

	mockTopicPublisher := mock.NewMockTopicPublisher(gomock.NewController(t))
	mockPaymentProvider := mock.NewMockPaymentProvider(gomock.NewController(t))

	container.TopicPublisher = mockTopicPublisher
	container.PaymentProvider = mockPaymentProvider

	return &Setup{
		Container:           container,
		Application:         bootstrap.NewApp(container),
		MockTopicPublisher:  mockTopicPublisher,
		MockPaymentProvider: mockPaymentProvider,
	}
}
