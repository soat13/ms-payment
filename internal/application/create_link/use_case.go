package create_link

import (
	"context"

	"github.com/google/uuid"
	"github.com/soat13/payment/internal/application/ports/out"
	"github.com/soat13/payment/internal/domain"
)

type (
	RequestPaymentLinkUseCase struct {
		repository      out.Repository
		publisher       out.TopicPublisher
		paymentProvider out.PaymentProvider
	}

	CreateLinkInput struct {
		ID uuid.UUID
	}
)

func NewRequestPaymentLinkUseCase(repository out.Repository, publisher out.TopicPublisher, service out.PaymentProvider) *RequestPaymentLinkUseCase {
	return &RequestPaymentLinkUseCase{
		repository:      repository,
		publisher:       publisher,
		paymentProvider: service,
	}
}

func (uc *RequestPaymentLinkUseCase) Execute(ctx context.Context, input CreateLinkInput) (*domain.Payment, error) {
	payment, err := uc.repository.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if err := uc.requestLink(ctx, payment); err != nil {
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

func (uc *RequestPaymentLinkUseCase) requestLink(ctx context.Context, payment *domain.Payment) error {
	input := out.CreateLinkInput{
		PaymentID:   payment.ID,
		Amount:      payment.Amount,
		Description: payment.Description,
	}

	output, err := uc.paymentProvider.CreateLink(ctx, input)
	if err != nil {
		return err
	}

	if err := payment.StartProcessing(output.Link, output.Provider, output.ProviderID); err != nil {
		return err
	}

	return nil
}
