package sns

import (
	"context"

	"github.com/soat13/oficina-utils/pkg/messaging"
	"github.com/soat13/payment/internal/application/ports/out"
	"github.com/soat13/payment/internal/domain"
)

type (
	Publisher struct {
		publisher messaging.TopicPublisher
	}
)

func NewPublisher(publisher messaging.TopicPublisher) out.TopicPublisher {
	return &Publisher{
		publisher: publisher,
	}
}

func (p Publisher) Publish(ctx context.Context, event domain.Event) error {
	var groupID *string

	if g, ok := event.(domain.GroupedEvent); ok {
		groupID = g.GroupID()
	}

	return p.publisher.Publish(ctx, messaging.TopicMessage{
		EventName: event.Name(),
		Payload:   event,
		GroupID:   groupID,
	})
}
