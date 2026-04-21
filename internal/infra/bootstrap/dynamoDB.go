package bootstrap

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func newDynamoDBClient(ctx context.Context, envs *Envs) (*dynamodb.Client, error) {
	cfg, err := getDynamoDBConfig(ctx, envs)
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(cfg), nil
}

func getDynamoDBConfig(ctx context.Context, envs *Envs) (aws.Config, error) {
	if !envs.IsTest {
		return config.LoadDefaultConfig(ctx)
	}

	return config.LoadDefaultConfig(
		ctx,
		config.WithRegion(envs.AwsRegion),
		config.WithBaseEndpoint(envs.AwsEndpoint),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", "test"),
		),
	)
}
