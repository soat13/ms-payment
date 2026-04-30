package create_payment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/ms-payment/internal/application/create_payment"
	"github.com/soat13/ms-payment/internal/application/ports/out/mock"
	"github.com/soat13/ms-payment/internal/domain"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type (
	deps struct {
		repository        *mock.MockRepository
		publisher         *mock.MockTopicPublisher
		useCase           *create_payment.CreatePaymentUseCase
		externalID        uuid.UUID
		amount            money.Money
		expectedPaymentID uuid.UUID
	}
)

func TestCreatePaymentUseCaseExecute(t *testing.T) {
	t.Parallel()

	t.Run("should create a pending payment and assign ID via repository", func(t *testing.T) {
		ctx := context.Background()
		deps := getDeps(t)

		mockRepositorySaveSuccess(t, ctx, deps)

		deps.publisher.EXPECT().
			Publish(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, event domain.StatusChangedEvent) error {
				if event.ID != deps.expectedPaymentID {
					t.Errorf("unexpected Payment ID in published event: got %v want %v", event.ID, deps.expectedPaymentID)
				}

				if event.Status != domain.StatusPending {
					t.Errorf("unexpected Payment Status in published event: got %v want %v", event.Status, domain.StatusPending)
				}

				if event.ExternalID != deps.externalID {
					t.Errorf("unexpected ExternalID: got %v want %v", event.ExternalID, deps.externalID)
				}

				return nil
			})

		payment, err := deps.useCase.Execute(ctx, create_payment.CreatePaymentInput{
			ExternalID: deps.externalID,
			Amount:     deps.amount,
		})

		require.NoError(t, err)
		require.Equal(t, deps.expectedPaymentID, payment.ID)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		ctx := context.Background()
		deps := getDeps(t)

		deps.repository.EXPECT().
			Save(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, payment *domain.Payment) error {
				return domain.ErrPaymentAlreadyExists
			})

		payment, err := deps.useCase.Execute(ctx, create_payment.CreatePaymentInput{
			ExternalID: deps.externalID,
			Amount:     deps.amount,
		})

		require.Equal(t, domain.ErrPaymentAlreadyExists, err)
		require.Nil(t, payment)
	})

	t.Run("should return error when publisher fails", func(t *testing.T) {
		ctx := context.Background()
		deps := getDeps(t)

		mockRepositorySaveSuccess(t, ctx, deps)

		deps.publisher.EXPECT().
			Publish(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, event domain.StatusChangedEvent) error {
				return errors.New("error publishing payment")
			})

		payment, err := deps.useCase.Execute(ctx, create_payment.CreatePaymentInput{
			ExternalID: deps.externalID,
			Amount:     deps.amount,
		})

		require.Equal(t, err, errors.New("error publishing payment"))
		require.Nil(t, payment)
	})
}

func getDeps(t *testing.T) deps {
	t.Helper()

	repository := mock.NewMockRepository(gomock.NewController(t))
	publisher := mock.NewMockTopicPublisher(gomock.NewController(t))
	amount, _ := money.New(12345)
	expectedPaymentID, _ := uuid.NewV7()

	return deps{
		repository:        repository,
		publisher:         publisher,
		useCase:           create_payment.NewCreatePaymentUseCase(repository, publisher),
		externalID:        uuid.New(),
		amount:            amount,
		expectedPaymentID: expectedPaymentID,
	}
}

func mockRepositorySaveSuccess(t *testing.T, ctx context.Context, d deps) {
	d.repository.EXPECT().
		Save(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, payment *domain.Payment) error {
			if payment.ExternalID != d.externalID {
				t.Errorf("unexpected ExternalID: got %v want %v", payment.ExternalID, d.externalID)
			}

			if payment.Amount != d.amount {
				t.Errorf("unexpected Amount: got %v want %v", payment.Amount, d.amount)
			}

			if payment.Status != domain.StatusPending {
				t.Errorf("unexpected Status: got %v want %v", payment.Status, domain.StatusPending)
			}

			if payment.ID != uuid.Nil {
				t.Errorf("payment ID should be nil before repository persistence")
			}

			payment.ID = d.expectedPaymentID

			return nil
		})
}
