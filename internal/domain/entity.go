package domain

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/entity"
	"github.com/soat13/oficina-utils/pkg/money"
)

type (
	Provider string

	Payment struct {
		entity.Timestamps
		ID                uuid.UUID
		ExternalID        uuid.UUID
		Amount            money.Money
		Status            Status
		Provider          *Provider
		ProviderPaymentID *string
		FailureReason     *string
		Metadata          *json.RawMessage
	}
)

func NewPendingPayment(externalID uuid.UUID, amount money.Money) (*Payment, error) {
	paymentEntity := &Payment{
		ExternalID: externalID,
		Amount:     amount,
		Status:     StatusPending,
	}

	if paymentEntity.Amount.Cents == 0 {
		return nil, ErrInvalidAmount
	}

	return paymentEntity, nil
}
