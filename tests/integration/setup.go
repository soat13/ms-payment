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
	Application    *bootstrap.App
	TopicPublisher *mock.MockPublisher
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

	mockedTopicPublisher := mock.NewMockPublisher(gomock.NewController(t))

	application := bootstrap.NewApp(
		container.Repository,
		mockedTopicPublisher,
		container.QueueSender,
		container.Consumer,
	)

	return &Setup{
		Application:    application,
		TopicPublisher: mockedTopicPublisher,
	}
}
