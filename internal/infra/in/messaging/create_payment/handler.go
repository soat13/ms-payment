package create_payment

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/messaging"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/payment/internal/application/create_payment"
	"github.com/soat13/payment/internal/domain"
)

type (
	Payload struct {
		ID     uuid.UUID   `json:"id"`
		Amount money.Money `json:"amount"`
	}
)

func Handler(useCase *create_payment.CreatePaymentUseCase) func(ctx context.Context, msg messaging.Message) error {
	return func(ctx context.Context, msg messaging.Message) error {
		payload := new(Payload)

		if err := msg.DecodePayload(payload); err != nil {
			return err
		}

		_, err := useCase.Execute(ctx, create_payment.CreatePaymentInput{
			ExternalID: payload.ID,
			Amount:     payload.Amount,
		})

		if errors.Is(err, domain.ErrPaymentAlreadyExists) {
			log.Println("Payment already exists", "external_id", payload.ID)
			return nil
		}

		return err
	}
}
