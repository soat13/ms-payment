package mercado_pago

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/merchantorder"
	"github.com/mercadopago/sdk-go/pkg/preference"
	"github.com/rs/zerolog/log"
	"github.com/soat13/ms-payment/internal/application/ports/out"
	"github.com/soat13/ms-payment/internal/domain"
)

type (
	MercadoPago struct {
		token               string
		webhookUrl          string
		preferenceClient    preference.Client
		merchantOrderClient merchantorder.Client
	}
)

var ErrPaymentNotYetProcessed = errors.New("payment not yet processed")

func NewMercadoPago(token string, webhookUrl string) *MercadoPago {
	cfg, err := config.New(token)
	if err != nil {
		panic("Failed to create MercadoPago config: " + err.Error())
	}

	return &MercadoPago{
		token:               token,
		webhookUrl:          webhookUrl,
		preferenceClient:    preference.NewClient(cfg),
		merchantOrderClient: merchantorder.NewClient(cfg),
	}
}

func (m *MercadoPago) FindByMerchantID(ctx context.Context, merchantID int) (*ProcessPaymentStatusResponse, error) {
	response, err := m.merchantOrderClient.Get(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	paymentID := uuid.MustParse(response.ExternalReference)
	payments := response.Payments

	if len(payments) == 0 {
		return nil, ErrPaymentNotYetProcessed
	}

	lastPayment := payments[len(payments)-1]

	log.Info().Str("payment_id", paymentID.String()).
		Str("mercado_pago_status", lastPayment.Status).
		Msg("Fetched payment status from Mercado Pago")

	return &ProcessPaymentStatusResponse{
		PaymentID: paymentID,
		Status:    mapStatus(lastPayment.Status),
	}, nil
}

func (m *MercadoPago) CreateLink(ctx context.Context, input out.CreateLinkInput) (*out.CreateLinkOutput, error) {
	request := m.createRequest(input)

	response, err := m.preferenceClient.Create(ctx, request)
	if err != nil {
		return nil, err
	}

	return &out.CreateLinkOutput{
		Link:       domain.Link(response.InitPoint),
		Provider:   domain.MercadoPagoProviderName,
		ProviderID: domain.ProviderID(response.ID),
	}, nil
}

func (m *MercadoPago) createRequest(input out.CreateLinkInput) preference.Request {
	item := preference.ItemRequest{
		Title:     input.Description,
		Quantity:  1,
		UnitPrice: float64(input.Amount.Cents) / 100,
	}

	return preference.Request{
		Items:             []preference.ItemRequest{item},
		ExternalReference: input.PaymentID.String(),
		NotificationURL:   m.webhookUrl,
	}
}

func mapStatus(mpStatus string) domain.Status {
	var statusMap = map[string]domain.Status{
		"approved":     domain.StatusSucceeded,
		"pending":      domain.StatusProcessing,
		"authorized":   domain.StatusProcessing,
		"cancelled":    domain.StatusFailed,
		"canceled":     domain.StatusFailed,
		"refunded":     domain.StatusFailed,
		"charged_back": domain.StatusFailed,
	}

	status, ok := statusMap[mpStatus]

	log.Info().
		Str("mercado_pago_status", mpStatus).
		Interface("mapped_status", status).
		Interface("mapping_exists", ok).
		Msg("Mapped Mercado Pago status to internal status")

	if !ok {
		status = domain.StatusError
	}

	log.Info().
		Str("mercado_pago_status", mpStatus).
		Interface("mapped_status", status).
		Interface("mapping_exists", ok).
		Msg("Depois do if")

	return status
}
