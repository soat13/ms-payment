package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/payment/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestNewPendingPayment(t *testing.T) {
	t.Parallel()

	t.Run("should return an error when amount is zero", func(t *testing.T) {
		amount, _ := money.New(0)
		payment, err := domain.NewPendingPayment(uuid.New(), amount)

		require.ErrorIs(t, err, domain.ErrInvalidAmount)
		require.Nil(t, payment)
	})

	t.Run("should create a new pending payment with valid input", func(t *testing.T) {
		externalID := uuid.New()
		amount, _ := money.New(123456)

		payment, err := domain.NewPendingPayment(externalID, amount)

		require.NoError(t, err)
		require.Equal(t, externalID, payment.ExternalID)
		require.Equal(t, amount.Cents, payment.Amount.Cents)
	})
}
