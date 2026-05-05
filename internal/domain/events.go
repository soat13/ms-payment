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
	PaymentURL *string   `json:"payment_url"`
}

func (e StatusChangedEvent) Name() string {
	return "payment-status-changed.fifo"
}

func NewStatusChangedEvent(payment Payment) Event {
	var url string
	if payment.Link != nil {
		url = string(*payment.Link)
	}

	return StatusChangedEvent{
		ID:         payment.ID,
		ExternalID: payment.ExternalID,
		Status:     payment.Status,
		PaymentURL: &url,
	}
}

func (e StatusChangedEvent) GroupID() *string {
	return new(e.ExternalID.String())
}
