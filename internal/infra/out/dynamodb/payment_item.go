package dynamodb

import (
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
	Description       string  `dynamodbav:"description"`
	ProviderPaymentID *string `dynamodbav:"provider_payment_id,omitempty"`
	FailureReason     *string `dynamodbav:"failure_reason,omitempty"`
	Version           int     `dynamodbav:"version"`
	Link              *string `dynamodbav:"link,omitempty"`
	CreatedAt         string  `dynamodbav:"created_at"`
	UpdatedAt         string  `dynamodbav:"updated_at"`
	GSI1PK            string  `dynamodbav:"gsi1pk"`
	GSI1SK            string  `dynamodbav:"gsi1sk"`
}

func toItem(p domain.Payment) (*paymentItem, error) {
	var provider *string
	if p.Provider != nil {
		provider = new(string)
		*provider = string(*p.Provider)
	}

	var providerPaymentID *string
	if p.ProviderPaymentID != nil {
		providerPaymentID = new(string(*p.ProviderPaymentID))
	}

	var link *string
	if p.Link != nil {
		link = new(string(*p.Link))
	}

	return &paymentItem{
		PK:                getPaymentPKByExternalID(p.ExternalID),
		SK:                getPaymentSKByExternalID(p.ExternalID),
		ID:                p.ID.String(),
		ExternalID:        p.ExternalID.String(),
		Status:            string(p.Status),
		Provider:          provider,
		AmountCents:       p.Amount.Cents,
		Description:       p.Description,
		ProviderPaymentID: providerPaymentID,
		Version:           p.Version,
		Link:              link,
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

	link := (*domain.Link)(i.Link)

	payment := &domain.Payment{
		ID:                id,
		ExternalID:        externalID,
		Amount:            amount,
		Link:              link,
		Description:       i.Description,
		Version:           i.Version,
		Status:            domain.Status(i.Status),
		ProviderPaymentID: (*domain.ProviderID)(i.ProviderPaymentID),
		Timestamps:        entity.NewTimestamps(createdAt, updatedAt),
	}

	if i.Provider != nil {
		payment.Provider = new(domain.ProviderName(*i.Provider))
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
