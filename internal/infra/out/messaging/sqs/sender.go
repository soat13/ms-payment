package sqs

import (
	"context"

	"github.com/soat13/oficina-utils/pkg/messaging"
	"github.com/soat13/ms-payment/internal/application/ports/out"
	"github.com/soat13/ms-payment/internal/domain"
	messagingHelper "github.com/soat13/ms-payment/internal/infra/out/messaging"
)

type Sender struct {
	Sender messaging.QueueSender
}

func NewSender(sender messaging.QueueSender) out.QueueSender {
	return &Sender{
		Sender: sender,
	}
}

func (s Sender) Send(ctx context.Context, event domain.Event) error {
	return s.Sender.Send(ctx, messaging.QueueMessage{
		EventName: event.Name(),
		Payload:   event,
		GroupID:   messagingHelper.ExtractGroupID(event),
	})
}
