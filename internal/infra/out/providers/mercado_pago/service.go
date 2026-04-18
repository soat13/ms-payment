package mercado_pago

import (
	"context"

	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/preference"
	"github.com/soat13/payment/internal/application/ports/out"
	"github.com/soat13/payment/internal/domain"
)

type (
	mercadoPago struct {
		token            string
		webhookUrl       string
		preferenceClient preference.Client
	}
)

func NewMercadoPago(token string, webhookUrl string) out.PaymentProvider {
	cfg, err := config.New(token)
	if err != nil {
		panic("Failed to create mercadoPago config: " + err.Error())
	}

	return &mercadoPago{
		token:            token,
		webhookUrl:       webhookUrl,
		preferenceClient: preference.NewClient(cfg),
	}
}

func (m *mercadoPago) CreateLink(ctx context.Context, input out.CreateLinkInput) (*out.CreateLinkOutput, error) {
	request := m.createRequest(input)

	response, err := m.preferenceClient.Create(ctx, request)
	if err != nil {
		return nil, err
	}

	return &out.CreateLinkOutput{
		Link:       domain.Link(response.SandboxInitPoint),
		Provider:   domain.MercadoPagoProviderName,
		ProviderID: domain.ProviderID(response.ID),
	}, nil
}

func (m *mercadoPago) createRequest(input out.CreateLinkInput) preference.Request {
	item := preference.ItemRequest{
		Title:     input.Description,
		Quantity:  1,
		UnitPrice: float64(input.Amount.Cents / 100),
	}

	return preference.Request{
		Items:             []preference.ItemRequest{item},
		ExternalReference: input.PaymentID.String(),
		BackURLs: &preference.BackURLsRequest{
			Success: m.webhookUrl,
			Pending: m.webhookUrl,
			Failure: m.webhookUrl,
		},
		NotificationURL: m.webhookUrl,
	}

}
