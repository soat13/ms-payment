package bootstrap

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/soat13/oficina-utils/pkg/messaging"
	"github.com/soat13/oficina-utils/pkg/messaging/sqs"
	app "github.com/soat13/payment/internal"
	"github.com/soat13/payment/internal/application/ports/out"
	infraDynamoDB "github.com/soat13/payment/internal/infra/out/dynamodb"
)

type (
	Envs struct {
		AwsEndpoint   string
		AwsBaseUrl    string
		AwsRegion     string
		DynamodbTable string
		DynamodbGSI   string
		IsTest        bool
	}

	Container struct {
		Envs Envs
	}
)

var (
	DefaultEnv = new(Envs)
)

func NewContainer(envs *Envs) app.Container {
	if envs != nil {
		return Container{
			Envs: *envs,
		}
	}

	return Container{
		Envs: getEnvs(),
	}
}

func (c Container) GetBroker(ctx context.Context) (messaging.Broker, error) {
	if c.Envs.IsTest {
		return sqs.NewSyncBroker(), nil
	}

	return sqs.NewBroker(ctx, c.Envs.AwsEndpoint, c.Envs.AwsBaseUrl)
}

func (c Container) GetRepository(ctx context.Context) (out.Repository, error) {
	client, err := newDynamoDBClient(ctx, c.Envs)
	if err != nil {
		return nil, err
	}

	return infraDynamoDB.NewRepository(client, c.Envs.DynamodbTable, c.Envs.DynamodbGSI), nil
}

func getEnvs() Envs {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("error loading .env: %v", err)
	}

	return Envs{
		AwsEndpoint:   os.Getenv("AWS_ENDPOINT"),
		AwsBaseUrl:    os.Getenv("AWS_BASE_URL"),
		DynamodbTable: os.Getenv("DYNAMODB_TABLE"),
		DynamodbGSI:   os.Getenv("DYNAMODB_GSI1"),
	}
}
