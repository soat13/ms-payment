package testkit

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/soat13/payment/internal/application/ports/out"
	infraDynamodb "github.com/soat13/payment/internal/infra/out/dynamodb"
	"github.com/stretchr/testify/require"
)

type IntegrationSetup struct {
	Client     *dynamodb.Client
	TableName  string
	IndexName  string
	Repository out.Repository
}

func NewIntegrationSetup(t *testing.T) *IntegrationSetup {
	t.Helper()

	ctx := context.Background()

	tableName := Getenv("DYNAMODB_TABLE_TEST", "payments_test")
	indexName := Getenv("DYNAMODB_GSI1", "gsi1")

	client := NewDynamoClient(t, ctx)
	repository := infraDynamodb.NewRepository(client, tableName, indexName)

	return &IntegrationSetup{
		Client:     client,
		TableName:  tableName,
		IndexName:  indexName,
		Repository: repository,
	}
}

func NewDynamoClient(t *testing.T, ctx context.Context) *dynamodb.Client {
	t.Helper()

	region := Getenv("AWS_REGION", "us-east-1")
	endpoint := Getenv("AWS_ENDPOINT_URL", "http://localhost:4566")

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", "test"),
		),
	)
	require.NoError(t, err)

	return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
}

func Getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
