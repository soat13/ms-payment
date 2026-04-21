package domain

import "github.com/google/uuid"

type Event interface {
	Name() string
}

type GroupedEvent interface {
	GroupID() *string
}

type StatusChangedEvent struct {
	ID         uuid.UUID `json:"id"`
	ExternalID uuid.UUID `json:"external_id"`
	Status     Status    `json:"status"`
}

func (e StatusChangedEvent) Name() string {
	return "payment-status-changed"
}

func NewStatusChangedEvent(payment Payment) Event {
	return StatusChangedEvent{
		ID:         payment.ID,
		ExternalID: payment.ExternalID,
		Status:     payment.Status,
	}
}
