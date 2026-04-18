package out

import (
	"context"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/payment/internal/domain"
)

type (
	CreateLinkOutput struct {
		Link       domain.Link
		Provider   domain.ProviderName
		ProviderID domain.ProviderID
	}

	CreateLinkInput struct {
		PaymentID   uuid.UUID
		Amount      money.Money
		Description string
	}

	PaymentProvider interface {
		CreateLink(ctx context.Context, input CreateLinkInput) (*CreateLinkOutput, error)
	}
)
