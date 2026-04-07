package create_payment

import (
	"context"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/payment/internal/application/ports/out"
	"github.com/soat13/payment/internal/domain"
)

type (
	CreatePaymentUseCase struct {
		repository out.Repository
	}

	CreatePaymentInput struct {
		ExternalID uuid.UUID
		Amount     money.Money
	}
)

func NewCreatePaymentUseCase(repository out.Repository) *CreatePaymentUseCase {
	return &CreatePaymentUseCase{repository: repository}
}

func (uc *CreatePaymentUseCase) Execute(ctx context.Context, input CreatePaymentInput) (*domain.Payment, error) {
	payment, err := domain.NewPendingPayment(input.ExternalID, input.Amount)
	if err != nil {
		return nil, err
	}

	if err := uc.repository.Create(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}
