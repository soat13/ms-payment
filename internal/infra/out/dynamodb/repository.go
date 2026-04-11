package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"github.com/soat13/payment/internal/application/ports/out"
	"github.com/soat13/payment/internal/domain"
)

type Repository struct {
	client    *dynamodb.Client
	tableName string
	indexName string
}

func NewRepository(client *dynamodb.Client, tableName, indexName string) out.Repository {
	return &Repository{
		client:    client,
		tableName: tableName,
		indexName: indexName,
	}
}

func (r *Repository) Create(ctx context.Context, payment *domain.Payment) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	now := time.Now()
	payment.ID = id
	payment.CreatedAt = now
	payment.UpdatedAt = now

	item, err := toItem(*payment)
	if err != nil {
		return err
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.tableName,
		Item:                av,
		ConditionExpression: new("attribute_not_exists(pk)"),
	})
	if err != nil {
		if _, ok := errors.AsType[*types.ConditionalCheckFailedException](err); ok {
			return domain.ErrPaymentAlreadyExists
		}

		return fmt.Errorf("dynamodb repository create: %w", err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	queryOutput, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              &r.indexName,
		KeyConditionExpression: new("gsi1pk = :gsi1pk AND gsi1sk = :gsi1sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi1pk": &types.AttributeValueMemberS{Value: getPaymentIDGSI1PK(id)},
			":gsi1sk": &types.AttributeValueMemberS{Value: getPaymentIDGSI1SK()},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb repository get by id: %w", err)
	}

	if len(queryOutput.Items) == 0 {
		return nil, domain.ErrPaymentNotFound
	}

	var item paymentItem
	if err := attributevalue.UnmarshalMap(queryOutput.Items[0], &item); err != nil {
		return nil, fmt.Errorf("dynamodb repository get by id unmarshal: %w", err)
	}

	return item.toDomain()
}
