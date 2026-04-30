package out

import (
	"context"

	"github.com/soat13/ms-payment/internal/domain"
)

type QueueSender interface {
	Send(ctx context.Context, event domain.Event) error
}
