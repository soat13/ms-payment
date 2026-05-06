package mercado_pago

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/soat13/ms-payment/internal/application/process_payment_status"
	"github.com/soat13/ms-payment/internal/domain"
	"github.com/soat13/ms-payment/internal/infra/out/providers/mercado_pago"
)

type (
	MercadoPagoHandler struct {
		mercadoPagoClient mercado_pago.Client
		useCase           process_payment_status.ProcessPaymentStatusUseCase
	}

	WebhookRequest struct {
		Resource string `json:"resource"`
		Topic    string `json:"topic"`
	}
)

var ErrInvalidResource = errors.New("invalid resource")

func NewHandler(client mercado_pago.Client, useCase process_payment_status.ProcessPaymentStatusUseCase) *MercadoPagoHandler {
	return &MercadoPagoHandler{
		mercadoPagoClient: client,
		useCase:           useCase,
	}
}

func (h *MercadoPagoHandler) Handle(c *fiber.Ctx) error {
	var input WebhookRequest
	if err := c.BodyParser(&input); err != nil {
		log.Warn().Err(err).
			Str("body", string(c.Body())).
			Msg("Invalid Mercado Pago webhook payload")

		return fiber.NewError(fiber.StatusBadRequest, "invalid webhook payload")
	}

	if input.shouldIgnore() {
		log.Debug().
			Str("body", string(c.Body())).
			Msg("Ignoring unsupported Mercado Pago webhook payload")

		return c.SendStatus(fiber.StatusNoContent)
	}

	merchantOrderID, err := input.extractMerchantOrderID()
	if err != nil {
		log.Warn().
			Err(err).
			Str("resource", input.Resource).
			Msg("Invalid Mercado Pago merchant order resource")

		return fiber.NewError(fiber.StatusBadRequest, ErrInvalidResource.Error())
	}

	processPaymentStatusResponse, err := h.mercadoPagoClient.FindByMerchantID(c.Context(), *merchantOrderID)
	if err != nil {
		return h.errorHandle(err)
	}

	useCaseInput := process_payment_status.ProcessPaymentStatusInput{
		ProviderPaymentID: processPaymentStatusResponse.PaymentID,
		NewStatus:         processPaymentStatusResponse.Status,
	}

	if err := h.useCase.Execute(c.Context(), useCaseInput); err != nil {
		return h.errorHandle(err)
	}

	return c.SendStatus(fiber.StatusOK)
}

func (i *WebhookRequest) shouldIgnore() bool {
	return strings.TrimSpace(i.Resource) == "" ||
		strings.TrimSpace(i.Topic) == "" ||
		i.Topic != "merchant_order"
}

func (i *WebhookRequest) extractMerchantOrderID() (*int, error) {
	resource := strings.TrimSpace(i.Resource)
	if resource == "" {
		return nil, ErrInvalidResource
	}

	parts := strings.Split(strings.TrimRight(resource, "/"), "/")
	lastPart := strings.TrimSpace(parts[len(parts)-1])

	merchantOrderID, err := strconv.Atoi(lastPart)
	if err != nil {
		return nil, ErrInvalidResource
	}

	return &merchantOrderID, nil
}

func (h *MercadoPagoHandler) errorHandle(err error) error {
	log.Err(err).Msg("Error processing Mercado Pago webhook")

	if errors.Is(err, domain.ErrPaymentNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "payment not found")
	}

	if errors.Is(err, mercado_pago.ErrPaymentNotYetProcessed) {
		return fiber.NewError(fiber.StatusAccepted, "payment not yet processed")
	}

	return fiber.NewError(fiber.StatusInternalServerError, "internal server error")
}
