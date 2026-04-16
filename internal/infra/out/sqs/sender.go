package sqs

import (
	"context"

	"github.com/soat13/oficina-utils/pkg/messaging"
	"github.com/soat13/payment/internal/application/ports/out"
	"github.com/soat13/payment/internal/domain"
)

type (
	Sender struct {
		Sender messaging.QueueSender
	}
)

func NewSender(sender messaging.QueueSender) out.QueueSender {
	return &Sender{
		Sender: sender,
	}
}

func (s Sender) Send(ctx context.Context, event domain.Event) error {
	var groupID *string

	if g, ok := event.(domain.GroupedEvent); ok {
		groupID = g.GroupID()
	}

	return s.Sender.Send(ctx, messaging.QueueMessage{
		EventName: event.Name(),
		Payload:   event,
		GroupID:   groupID,
	})
}
