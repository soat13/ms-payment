package dynamodb

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/entity"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/payment/internal/domain"
)

type paymentItem struct {
	PK                string  `dynamodbav:"pk"`
	SK                string  `dynamodbav:"sk"`
	ID                string  `dynamodbav:"id"`
	ExternalID        string  `dynamodbav:"external_id"`
	Status            string  `dynamodbav:"status"`
	Provider          *string `dynamodbav:"provider,omitempty"`
	AmountCents       int64   `dynamodbav:"amount_cents"`
	ProviderPaymentID *string `dynamodbav:"provider_payment_id,omitempty"`
	FailureReason     *string `dynamodbav:"failure_reason,omitempty"`
	Metadata          *string `dynamodbav:"metadata,omitempty"`
	CreatedAt         string  `dynamodbav:"created_at"`
	UpdatedAt         string  `dynamodbav:"updated_at"`
	GSI1PK            string  `dynamodbav:"gsi1pk"`
	GSI1SK            string  `dynamodbav:"gsi1sk"`
}

func toItem(p domain.Payment) (*paymentItem, error) {
	var provider *string
	if p.Provider != nil {
		provider = new(string(*p.Provider))
	}

	var metadata *string
	if p.Metadata != nil {
		metadata = new(string(*p.Metadata))
	}

	return &paymentItem{
		PK:                getPaymentPKByExternalID(p.ExternalID),
		SK:                getPaymentSKByExternalID(p.ExternalID),
		ID:                p.ID.String(),
		ExternalID:        p.ExternalID.String(),
		Status:            string(p.Status),
		Provider:          provider,
		AmountCents:       p.Amount.Cents,
		ProviderPaymentID: p.ProviderPaymentID,
		FailureReason:     p.FailureReason,
		Metadata:          metadata,
		CreatedAt:         p.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:         p.UpdatedAt.Format(time.RFC3339Nano),
		GSI1PK:            getPaymentIDGSI1PK(p.ID),
		GSI1SK:            getPaymentIDGSI1SK(),
	}, nil
}

func (i paymentItem) toDomain() (*domain.Payment, error) {
	id, err := uuid.Parse(i.ID)
	if err != nil {
		return nil, err
	}

	externalID, err := uuid.Parse(i.ExternalID)
	if err != nil {
		return nil, err
	}

	amount, err := money.New(i.AmountCents)
	if err != nil {
		return nil, err
	}

	createdAt, err := time.Parse(time.RFC3339Nano, i.CreatedAt)
	if err != nil {
		return nil, err
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, i.UpdatedAt)
	if err != nil {
		return nil, err
	}

	payment := &domain.Payment{
		ID:                id,
		ExternalID:        externalID,
		Amount:            amount,
		Status:            domain.Status(i.Status),
		ProviderPaymentID: i.ProviderPaymentID,
		FailureReason:     i.FailureReason,
		Timestamps:        entity.NewTimestamps(createdAt, updatedAt),
	}

	if i.Provider != nil {
		payment.Provider = new(domain.Provider(*i.Provider))
	}

	if i.Metadata != nil {
		payment.Metadata = new(json.RawMessage(*i.Metadata))
	}

	return payment, nil
}

func getPaymentPKByExternalID(externalID uuid.UUID) string {
	return "PAYMENT#" + externalID.String()
}

func getPaymentSKByExternalID(externalID uuid.UUID) string {
	return "PAYMENT#" + externalID.String()
}

func getPaymentIDGSI1PK(id uuid.UUID) string {
	return "PAYMENT_ID#" + id.String()
}

func getPaymentIDGSI1SK() string {
	return "PAYMENT"
}
