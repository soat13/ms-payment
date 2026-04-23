package create_payment

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/soat13/oficina-utils/pkg/messaging"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/payment/internal/application/create_payment"
	"github.com/soat13/payment/internal/domain"
)

type (
	Payload struct {
		ID          uuid.UUID   `json:"id"`
		Amount      money.Money `json:"amount"`
		Description string      `json:"description"`
	}
)

func Handler(useCase *create_payment.CreatePaymentUseCase) func(ctx context.Context, msg messaging.Message) error {
	return func(ctx context.Context, msg messaging.Message) error {
		log.Info().Msg("Received message to create payment")

		payload := new(Payload)

		payload, err := messaging.DecodePayload[Payload](msg)

		if err != nil {
			log.Error().Err(err).Msg("Failed to decode payload")
			return err
		}

		_, err = useCase.Execute(ctx, create_payment.CreatePaymentInput{
			ExternalID:  payload.ID,
			Amount:      payload.Amount,
			Description: payload.Description,
		})

		if errors.Is(err, domain.ErrPaymentAlreadyExists) {
			log.Warn().Err(err).Msg("Payment already exists, skipping creation")
			return nil
		}

		return err
	}
}
