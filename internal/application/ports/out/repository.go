package out

import (
	"context"

	"github.com/google/uuid"
	"github.com/soat13/payment/internal/domain"
)

type (
	Repository interface {
		Create(ctx context.Context, payment *domain.Payment) error
		CountByExternalID(ctx context.Context, externalID uuid.UUID) (int, error)
		GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
		GetByExternalID(ctx context.Context, externalID uuid.UUID) (*domain.Payment, error)
	}
)
