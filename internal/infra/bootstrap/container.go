package bootstrap

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/soat13/oficina-utils/pkg/messaging"
	"github.com/soat13/oficina-utils/pkg/messaging/sns"
	"github.com/soat13/oficina-utils/pkg/messaging/sqs"
	"github.com/soat13/payment/internal/application/ports/out"
	infraDynamoDB "github.com/soat13/payment/internal/infra/out/dynamodb"
	infraSNSPublisher "github.com/soat13/payment/internal/infra/out/sns"
	infraSQSPublisher "github.com/soat13/payment/internal/infra/out/sqs"
)

type (
	Envs struct {
		AwsEndpoint   string
		AwsBaseSQSUrl string
		AwsBaseSNSARN string
		AwsRegion     string
		DynamodbTable string
		DynamodbGSI   string
		IsTest        bool
	}

	Container struct {
		Envs           Envs
		Consumer       messaging.Consumer
		TopicPublisher out.TopicPublisher
		QueueSender    out.QueueSender
		Repository     out.Repository
	}
)

var (
	DefaultEnvs = new(Envs)
)

func NewContainer(ctx context.Context, envs *Envs) (*Container, error) {
	if envs == nil {
		envs = new(getEnvs())
	}

	queueBroker, err := getQueueBroker(envs)
	if err != nil {
		return nil, err
	}

	snsPublisher, err := sns.NewPublisher(ctx, envs.AwsEndpoint, envs.AwsBaseSNSARN)
	if err != nil {
		return nil, err
	}

	client, err := newDynamoDBClient(ctx, envs)
	if err != nil {
		return nil, err
	}

	return &Container{
		Envs:           getEnvs(),
		Consumer:       queueBroker,
		TopicPublisher: infraSNSPublisher.NewPublisher(snsPublisher),
		QueueSender:    infraSQSPublisher.NewSender(queueBroker),
		Repository:     infraDynamoDB.NewRepository(client, envs.DynamodbTable, envs.DynamodbGSI),
	}, nil
}

func getEnvs() Envs {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("error loading .env: %v", err)
	}

	return Envs{
		AwsEndpoint:   os.Getenv("AWS_ENDPOINT"),
		AwsBaseSQSUrl: os.Getenv("AWS_BASE_URL"),
		AwsBaseSNSARN: os.Getenv("AWS_BASE_SNS_ARN"),
		DynamodbTable: os.Getenv("DYNAMODB_TABLE"),
		DynamodbGSI:   os.Getenv("DYNAMODB_GSI1"),
	}
}

func getQueueBroker(envs *Envs) (messaging.QueueBroker, error) {
	if envs.IsTest {
		return sqs.NewSyncBroker(), nil
	}

	return sqs.NewBroker(context.Background(), envs.AwsEndpoint, envs.AwsBaseSQSUrl)
}
