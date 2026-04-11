package out

import (
	"context"

	"github.com/google/uuid"
	"github.com/soat13/payment/internal/domain"
)

type (
	Repository interface {
		Create(ctx context.Context, payment *domain.Payment) error
		GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	}
)
