package create_payment_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/payment/internal/application/create_payment"
	"github.com/soat13/payment/internal/domain"
	"github.com/soat13/payment/tests/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePaymentUseCase(t *testing.T) {
	t.Parallel()

	t.Run("should reject duplicate external_id", func(t *testing.T) {
		ctx := context.Background()
		setup := testkit.NewIntegrationSetup(t)
		useCase := create_payment.NewCreatePaymentUseCase(setup.Repository)

		externalID := uuid.New()
		amount, _ := money.New(12345)

		// given
		_, err := useCase.Execute(ctx, create_payment.CreatePaymentInput{
			ExternalID: externalID,
			Amount:     amount,
		})
		require.NoError(t, err)

		// when
		_, err = useCase.Execute(ctx, create_payment.CreatePaymentInput{
			ExternalID: externalID,
			Amount:     amount,
		})

		// then
		require.ErrorIs(t, err, domain.ErrPaymentAlreadyExists)
	})

	t.Run("should create payment successfully", func(t *testing.T) {

		ctx := context.Background()
		setup := testkit.NewIntegrationSetup(t)
		useCase := create_payment.NewCreatePaymentUseCase(setup.Repository)

		externalID := uuid.New()
		amount, _ := money.New(12345)

		// when
		payment, err := useCase.Execute(ctx, create_payment.CreatePaymentInput{
			ExternalID: externalID,
			Amount:     amount,
		})

		// then
		require.NoError(t, err)
		require.NotNil(t, payment)
		assert.NotEqual(t, uuid.Nil, payment.ID)
		assert.Equal(t, externalID, payment.ExternalID)
		assert.Equal(t, amount.Cents, payment.Amount.Cents)
		assert.Equal(t, domain.StatusPending, payment.Status)

		found, err := setup.Repository.GetByID(ctx, payment.ID)
		require.NoError(t, err)

		require.Equal(t, payment.ID, found.ID)
		require.Equal(t, payment.ExternalID, found.ExternalID)
		require.Equal(t, payment.Amount.Cents, found.Amount.Cents)
		require.Equal(t, payment.Status, found.Status)
	})
}
