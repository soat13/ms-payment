package out

import (
	"context"

	"github.com/soat13/ms-payment/internal/domain"
)

type TopicPublisher interface {
	Publish(ctx context.Context, event domain.Event) error
}
