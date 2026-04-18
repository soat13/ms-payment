package bootstrap

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/soat13/oficina-utils/pkg/awsconfig"
	"github.com/soat13/oficina-utils/pkg/messaging"
	"github.com/soat13/oficina-utils/pkg/messaging/sns"
	"github.com/soat13/oficina-utils/pkg/messaging/sqs"
	"github.com/soat13/payment/internal/application/ports/out"
	infraDynamoDB "github.com/soat13/payment/internal/infra/out/dynamodb"
	infraSNSPublisher "github.com/soat13/payment/internal/infra/out/messaging/sns"
	infraSQSPublisher "github.com/soat13/payment/internal/infra/out/messaging/sqs"
	"github.com/soat13/payment/internal/infra/out/providers/mercado_pago"
)

type (
	Envs struct {
		AwsEndpoint           string
		AwsSecretAccessKey    string
		AwsAccessKeyId        string
		AwsSessionToken       string
		AwsBaseSQSUrl         string
		AwsBaseSNSARN         string
		AwsRegion             string
		DynamodbTable         string
		DynamodbGSI           string
		IsTest                bool
		MercadoPagoToken      string
		MercadoPagoWebhookUrl string
	}

	Container struct {
		Consumer        messaging.Consumer
		TopicPublisher  out.TopicPublisher
		QueueSender     out.QueueSender
		Repository      out.Repository
		PaymentProvider out.PaymentProvider
	}
)

func NewContainer(ctx context.Context, envs *Envs) (*Container, error) {
	if envs == nil {
		envs = new(getEnvs())
	}

	awsConfig := getAwsConfig(envs)

	queueBroker, err := getQueueBroker(envs, awsConfig)
	if err != nil {
		return nil, err
	}

	snsPublisher, err := sns.NewPublisher(ctx, awsConfig, envs.AwsBaseSNSARN)
	if err != nil {
		return nil, err
	}

	client, err := newDynamoDBClient(ctx, envs)
	if err != nil {
		return nil, err
	}

	return &Container{
		Consumer:        queueBroker,
		TopicPublisher:  infraSNSPublisher.NewPublisher(snsPublisher),
		QueueSender:     infraSQSPublisher.NewSender(queueBroker),
		Repository:      infraDynamoDB.NewRepository(client, envs.DynamodbTable, envs.DynamodbGSI),
		PaymentProvider: mercado_pago.NewMercadoPago(envs.MercadoPagoToken, envs.MercadoPagoWebhookUrl),
	}, nil
}

func getEnvs() Envs {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("error loading .env: %v", err)
	}

	return Envs{
		AwsEndpoint:           os.Getenv("AWS_ENDPOINT"),
		AwsSecretAccessKey:    os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AwsAccessKeyId:        os.Getenv("AWS_ACCESS_KEY_ID"),
		AwsSessionToken:       os.Getenv("AWS_SESSION_TOKEN"),
		AwsBaseSQSUrl:         os.Getenv("AWS_SQS_BASE_URL"),
		AwsBaseSNSARN:         os.Getenv("AWS_BASE_SNS_ARN"),
		DynamodbTable:         os.Getenv("DYNAMODB_TABLE"),
		DynamodbGSI:           os.Getenv("DYNAMODB_GSI1"),
		MercadoPagoToken:      os.Getenv("MERCADO_PAGO_TOKEN"),
		MercadoPagoWebhookUrl: os.Getenv("MERCADO_PAGO_WEBHOOK_URL"),
	}
}

func getQueueBroker(envs *Envs, config awsconfig.Config) (messaging.QueueBroker, error) {
	if envs.IsTest {
		return sqs.NewSyncBroker(), nil
	}

	return sqs.NewBroker(context.Background(), config, envs.AwsBaseSQSUrl)
}

func getAwsConfig(envs *Envs) awsconfig.Config {
	return awsconfig.Config{
		Region:          envs.AwsRegion,
		EndpointURL:     envs.AwsEndpoint,
		AccessKeyID:     envs.AwsAccessKeyId,
		SecretAccessKey: envs.AwsSecretAccessKey,
		SessionToken:    envs.AwsSessionToken,
	}
}
