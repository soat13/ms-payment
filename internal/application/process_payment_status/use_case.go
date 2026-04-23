package process_payment_status

import (
	"context"

	"github.com/google/uuid"
	"github.com/soat13/payment/internal/application/ports/out"
	"github.com/soat13/payment/internal/domain"
)

type (
	ProcessPaymentStatusUseCase struct {
		repository      out.Repository
		publisher       out.TopicPublisher
		paymentProvider out.PaymentProvider
	}

	ProcessPaymentStatusInput struct {
		ProviderPaymentID uuid.UUID
		NewStatus         domain.Status
	}
)

func NewProcessPaymentStatusUseCase(
	repository out.Repository,
	publisher out.TopicPublisher,
	paymentProvider out.PaymentProvider,
) *ProcessPaymentStatusUseCase {
	return &ProcessPaymentStatusUseCase{
		repository:      repository,
		publisher:       publisher,
		paymentProvider: paymentProvider,
	}
}

func (uc ProcessPaymentStatusUseCase) Execute(ctx context.Context, input ProcessPaymentStatusInput) error {
	payment, err := uc.repository.GetByID(ctx, input.ProviderPaymentID)
	if err != nil {
		return err
	}

	if err := payment.ApplyAttemptResult(input.NewStatus); err != nil {
		return err
	}

	if err := uc.repository.Save(ctx, payment); err != nil {
		return err
	}

	return uc.publisher.Publish(ctx, domain.NewStatusChangedEvent(*payment))
}
