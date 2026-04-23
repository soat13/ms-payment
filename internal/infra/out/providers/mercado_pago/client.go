package mercado_pago

import (
	"context"

	"github.com/google/uuid"
	"github.com/soat13/payment/internal/domain"
)

type (
	ProcessPaymentStatusResponse struct {
		PaymentID uuid.UUID
		Status    domain.Status
	}

	Client interface {
		FindByMerchantID(ctx context.Context, merchantID int) (*ProcessPaymentStatusResponse, error)
	}
)
