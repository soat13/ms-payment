package integration

import (
	"context"
	"testing"

	app "github.com/soat13/payment/internal"
	"github.com/soat13/payment/internal/infra/bootstrap"
	"github.com/stretchr/testify/require"
)

type Setup struct {
	Application *app.App
}

func NewIntegrationSetup(t *testing.T) *Setup {
	t.Helper()

	env := bootstrap.Envs{
		IsTest:        true,
		AwsRegion:     "us-east-1",
		AwsEndpoint:   "http://localhost:4566",
		DynamodbGSI:   "gsi1",
		DynamodbTable: "payments_test",
	}

	container := bootstrap.NewContainer(&env)

	ctx := context.Background()
	application, err := app.New(ctx, container)
	if err != nil {
		require.NoError(t, err)
	}

	return &Setup{
		Application: application,
	}
}
