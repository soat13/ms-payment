package create_payment_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/payment/internal/application/create_payment"
	"github.com/soat13/payment/internal/application/ports/out/mock"
	"github.com/soat13/payment/internal/domain"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreatePaymentUseCaseExecute(t *testing.T) {
	t.Parallel()

	t.Run("should return error when repository fails", func(t *testing.T) {
		ctx := context.Background()
		repository := mock.NewMockRepository(gomock.NewController(t))
		useCase := create_payment.NewCreatePaymentUseCase(repository)
		externalID := uuid.New()
		amount, _ := money.New(12345)

		repository.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, payment *domain.Payment) error {
				return domain.ErrPaymentAlreadyExists
			})

		payment, err := useCase.Execute(ctx, create_payment.CreatePaymentInput{
			ExternalID: externalID,
			Amount:     amount,
		})

		require.Equal(t, domain.ErrPaymentAlreadyExists, err)
		require.Nil(t, payment)
	})

	t.Run("should create a pending payment and assign ID via repository", func(t *testing.T) {
		ctx := context.Background()
		repository := mock.NewMockRepository(gomock.NewController(t))
		useCase := create_payment.NewCreatePaymentUseCase(repository)
		externalID := uuid.New()
		amount, _ := money.New(12345)
		expectedPaymentID, _ := uuid.NewV7()

		repository.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, payment *domain.Payment) error {
				if payment.ExternalID != externalID {
					t.Errorf("unexpected ExternalID: got %v want %v", payment.ExternalID, externalID)
				}

				if payment.Amount != amount {
					t.Errorf("unexpected Amount: got %v want %v", payment.Amount, amount)
				}

				if payment.Status != domain.StatusPending {
					t.Errorf("unexpected Status: got %v want %v", payment.Status, domain.StatusPending)
				}

				if payment.ID != uuid.Nil {
					t.Errorf("payment ID should be nil before repository persistence")
				}

				payment.ID = expectedPaymentID

				return nil
			})

		payment, err := useCase.Execute(ctx, create_payment.CreatePaymentInput{
			ExternalID: externalID,
			Amount:     amount,
		})

		require.NoError(t, err)
		require.Equal(t, expectedPaymentID, payment.ID)
	})
}
