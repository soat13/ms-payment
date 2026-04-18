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
		publisher  out.TopicPublisher
	}

	CreatePaymentInput struct {
		Description string
		ExternalID  uuid.UUID
		Amount      money.Money
	}
)

func NewCreatePaymentUseCase(repository out.Repository, publisher out.TopicPublisher) *CreatePaymentUseCase {
	return &CreatePaymentUseCase{
		repository: repository,
		publisher:  publisher,
	}
}

func (uc *CreatePaymentUseCase) Execute(ctx context.Context, input CreatePaymentInput) (*domain.Payment, error) {
	payment, err := domain.NewPendingPayment(input.ExternalID, input.Amount, input.Description)
	if err != nil {
		return nil, err
	}

	if err := uc.repository.Save(ctx, payment); err != nil {
		return nil, err
	}

	if err := uc.publisher.Publish(ctx, domain.NewStatusChangedEvent(*payment)); err != nil {
		return nil, err
	}

	return payment, nil
}
