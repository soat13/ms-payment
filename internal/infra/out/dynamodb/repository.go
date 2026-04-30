package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"github.com/soat13/ms-payment/internal/application/ports/out"
	"github.com/soat13/ms-payment/internal/domain"
)

type Repository struct {
	client    *dynamodb.Client
	tableName string
	indexName string
}

var ErrConcurrentModification = errors.New("concurrent modification detected")

func NewRepository(client *dynamodb.Client, tableName, indexName string) out.Repository {
	return &Repository{
		client:    client,
		tableName: tableName,
		indexName: indexName,
	}
}

func (r *Repository) Save(ctx context.Context, payment *domain.Payment) error {
	if payment.Version == 0 {
		return r.create(ctx, payment)
	}

	return r.update(ctx, payment)
}

func (r *Repository) CountByExternalID(ctx context.Context, externalID uuid.UUID) (int, error) {
	output, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		KeyConditionExpression: new("pk = :pk AND sk = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: getPaymentPKByExternalID(externalID)},
			":sk": &types.AttributeValueMemberS{Value: getPaymentSKByExternalID(externalID)},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("dynamodb repository count by external id: %w", err)
	}

	return len(output.Items), nil
}

func (r *Repository) GetByExternalID(ctx context.Context, externalID uuid.UUID) (*domain.Payment, error) {
	output, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.tableName,
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: getPaymentPKByExternalID(externalID)},
			"sk": &types.AttributeValueMemberS{Value: getPaymentSKByExternalID(externalID)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb repository get by external id: %w", err)
	}

	if len(output.Item) == 0 {
		return nil, domain.ErrPaymentNotFound
	}

	var item paymentItem
	if err := attributevalue.UnmarshalMap(output.Item, &item); err != nil {
		return nil, fmt.Errorf("dynamodb repository get by external id unmarshal: %w", err)
	}

	return item.toDomain()
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

func (r *Repository) create(ctx context.Context, payment *domain.Payment) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	payment.ID = id
	payment.CreatedAt = now
	payment.UpdatedAt = now
	payment.Version = 1

	item, err := toItem(*payment)
	if err != nil {
		return err
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(pk) AND attribute_not_exists(sk)"),
	})
	if err != nil {
		if _, ok := errors.AsType[*types.ConditionalCheckFailedException](err); ok {
			return domain.ErrPaymentAlreadyExists
		}

		return fmt.Errorf("dynamodb repository create payment: %w", err)
	}

	return nil
}
func (r *Repository) update(ctx context.Context, payment *domain.Payment) error {
	currentVersion := payment.Version
	nextVersion := currentVersion + 1
	updatedAt := time.Now().UTC()

	clone := *payment
	clone.UpdatedAt = updatedAt
	clone.Version = nextVersion

	item, err := toItem(clone)
	if err != nil {
		return err
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
		ConditionExpression: aws.String(
			"attribute_exists(pk) AND attribute_exists(sk) AND version = :version",
		),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":version": &types.AttributeValueMemberN{Value: strconv.Itoa(currentVersion)},
		},
	})
	if err != nil {
		if _, ok := errors.AsType[*types.ConditionalCheckFailedException](err); ok {
			return ErrConcurrentModification
		}

		return fmt.Errorf("dynamodb repository update: %w", err)
	}

	payment.UpdatedAt = updatedAt
	payment.Version = nextVersion
	return nil
}
