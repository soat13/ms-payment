package domain

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/entity"
	"github.com/soat13/oficina-utils/pkg/money"
)

type (
	Payment struct {
		entity.Timestamps
		ID                uuid.UUID
		ExternalID        uuid.UUID
		Amount            money.Money
		Description       string
		Status            Status
		Provider          *ProviderName
		ProviderPaymentID *ProviderID
		Metadata          *json.RawMessage
		Link              *Link
		Version           int
	}
)

var (
	ErrInvalidAmount           = errors.New("amount must be greater than zero")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrPaymentAlreadyExists    = errors.New("payment already exists")
	ErrPaymentNotFound         = errors.New("payment not found")
)

func NewPendingPayment(externalID uuid.UUID, amount money.Money, description string) (*Payment, error) {
	paymentEntity := &Payment{
		ExternalID:  externalID,
		Amount:      amount,
		Status:      StatusPending,
		Description: description,
		Version:     0,
	}

	if paymentEntity.Amount.Cents == 0 {
		return nil, ErrInvalidAmount
	}

	return paymentEntity, nil
}

func (p *Payment) IsPending() bool {
	return p.Status == StatusPending
}

func (p *Payment) IsProcessing() bool {
	return p.Status == StatusProcessing
}

func (p *Payment) IsFailed() bool {
	return p.Status == StatusFailed
}

func (p *Payment) StartProcessing(link Link, provider ProviderName, providerID ProviderID) error {
	if !p.IsPending() {
		return ErrInvalidStatusTransition
	}

	if err := link.Validate(); err != nil {
		return err
	}

	if err := provider.Validate(); err != nil {
		return err
	}

	if err := providerID.Validate(); err != nil {
		return err
	}

	p.Status = StatusProcessing
	p.Link = &link
	p.Provider = &provider
	p.ProviderPaymentID = &providerID

	return nil
}

func (p *Payment) ApplyAttemptResult(newStatus Status) error {
	if !newStatus.IsAttemptResult() {
		return ErrInvalidStatusTransition
	}

	if !p.IsProcessing() && !p.IsFailed() {
		return ErrInvalidStatusTransition
	}

	p.Status = newStatus

	return nil
}
