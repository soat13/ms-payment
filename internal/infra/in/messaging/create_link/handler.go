package create_link

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/soat13/oficina-utils/pkg/messaging"
	"github.com/soat13/payment/internal/application/create_link"
	"github.com/soat13/payment/internal/domain"
)

type (
	Payload struct {
		ID     uuid.UUID     `json:"id"`
		Status domain.Status `json:"status"`
	}
)

func Handler(useCase *create_link.RequestPaymentLinkUseCase) func(ctx context.Context, msg messaging.Message) error {
	return func(ctx context.Context, msg messaging.Message) error {
		log.Info().Msg("Received message to create payment link")

		payload, err := messaging.DecodePayload[Payload](msg)

		if err != nil {
			log.Error().Err(err).Msg("Received message to create payment link")
			return err
		}

		log.Debug().Interface("payload", payload).Msg("Received message to create payment link")

		if payload.Status != domain.StatusPending {
			log.Debug().Interface("status", payload.Status).Msg("Payment status is not pending, skipping link creation")
			return nil
		}

		_, err = useCase.Execute(ctx, create_link.CreateLinkInput{ID: payload.ID})

		if err != nil {
			log.Error().Err(err).Msg("Failed to create payment link")
		}

		return err
	}
}
