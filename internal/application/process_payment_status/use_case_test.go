package process_payment_status_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/ms-payment/internal/application/ports/out/mock"
	"github.com/soat13/ms-payment/internal/application/process_payment_status"
	"github.com/soat13/ms-payment/internal/domain"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type deps struct {
	repository      *mock.MockRepository
	publisher       *mock.MockTopicPublisher
	paymentProvider *mock.MockPaymentProvider
	useCase         *process_payment_status.ProcessPaymentStatusUseCase
	id              uuid.UUID
	payment         *domain.Payment
	newStatus       domain.Status
}

func TestProcessPaymentStatusUseCaseExecute(t *testing.T) {
	t.Parallel()

	t.Run("should process payment status and publish status changed event", func(t *testing.T) {
		ctx := context.Background()
		deps := getDeps(t)

		mockRepositoryGetByIDSuccess(t, ctx, deps)
		mockRepositorySaveSuccess(t, ctx, deps)
		mockPublisher(t, ctx, deps, nil)

		err := deps.useCase.Execute(ctx, process_payment_status.ProcessPaymentStatusInput{
			ProviderPaymentID: deps.id,
			NewStatus:         deps.newStatus,
		})

		require.NoError(t, err)
		require.Equal(t, deps.newStatus, deps.payment.Status)
	})

	t.Run("should return error when repository get by id fails", func(t *testing.T) {
		ctx := context.Background()
		deps := getDeps(t)

		expectedErr := errors.New("payment not found")

		deps.repository.EXPECT().
			GetByID(ctx, deps.id).
			Return(nil, expectedErr)

		err := deps.useCase.Execute(ctx, process_payment_status.ProcessPaymentStatusInput{
			ProviderPaymentID: deps.id,
			NewStatus:         deps.newStatus,
		})

		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("should return error when apply attempt result fails", func(t *testing.T) {
		ctx := context.Background()
		deps := getDeps(t)

		invalidStatus := domain.Status("invalid_status")

		mockRepositoryGetByIDSuccess(t, ctx, deps)

		err := deps.useCase.Execute(ctx, process_payment_status.ProcessPaymentStatusInput{
			ProviderPaymentID: deps.id,
			NewStatus:         invalidStatus,
		})

		require.Error(t, err)
	})

	t.Run("should return error when repository save fails", func(t *testing.T) {
		ctx := context.Background()
		deps := getDeps(t)

		expectedErr := errors.New("error saving payment")

		mockRepositoryGetByIDSuccess(t, ctx, deps)

		deps.repository.EXPECT().
			Save(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, payment *domain.Payment) error {
				if payment.ID != deps.payment.ID {
					t.Errorf("unexpected Payment ID: got %v want %v", payment.ID, deps.payment.ID)
				}

				if payment.Status != deps.newStatus {
					t.Errorf("unexpected Status: got %v want %v", payment.Status, deps.newStatus)
				}

				return expectedErr
			})

		err := deps.useCase.Execute(ctx, process_payment_status.ProcessPaymentStatusInput{
			ProviderPaymentID: deps.id,
			NewStatus:         deps.newStatus,
		})

		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("should return error when publisher fails", func(t *testing.T) {
		ctx := context.Background()
		deps := getDeps(t)

		expectedErr := errors.New("error publishing payment")

		mockRepositoryGetByIDSuccess(t, ctx, deps)
		mockRepositorySaveSuccess(t, ctx, deps)
		mockPublisher(t, ctx, deps, expectedErr)

		err := deps.useCase.Execute(ctx, process_payment_status.ProcessPaymentStatusInput{
			ProviderPaymentID: deps.id,
			NewStatus:         deps.newStatus,
		})

		require.ErrorIs(t, err, expectedErr)
	})
}

func getDeps(t *testing.T) deps {
	t.Helper()

	repository := mock.NewMockRepository(gomock.NewController(t))
	publisher := mock.NewMockTopicPublisher(gomock.NewController(t))
	paymentProvider := mock.NewMockPaymentProvider(gomock.NewController(t))

	amount, _ := money.New(12345)

	payment := &domain.Payment{
		ID:          uuid.New(),
		ExternalID:  uuid.New(),
		Amount:      amount,
		Description: "payment description",
		Status:      domain.StatusProcessing,
	}

	return deps{
		repository:      repository,
		publisher:       publisher,
		paymentProvider: paymentProvider,
		useCase: process_payment_status.NewProcessPaymentStatusUseCase(
			repository,
			publisher,
			paymentProvider,
		),
		id:        payment.ID,
		payment:   payment,
		newStatus: domain.StatusSucceeded,
	}
}

func mockRepositoryGetByIDSuccess(t *testing.T, ctx context.Context, d deps) {
	t.Helper()

	d.repository.EXPECT().
		GetByID(ctx, d.id).
		Return(d.payment, nil)
}

func mockRepositorySaveSuccess(t *testing.T, ctx context.Context, d deps) {
	t.Helper()

	d.repository.EXPECT().
		Save(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, payment *domain.Payment) error {
			if payment.ID != d.payment.ID {
				t.Errorf("unexpected Payment ID: got %v want %v", payment.ID, d.payment.ID)
			}

			if payment.Status != d.newStatus {
				t.Errorf("unexpected Status: got %v want %v", payment.Status, d.newStatus)
			}

			return nil
		})
}

func mockPublisher(t *testing.T, ctx context.Context, d deps, response error) {
	t.Helper()

	d.publisher.EXPECT().
		Publish(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, event domain.StatusChangedEvent) error {
			if event.ID != d.payment.ID {
				t.Errorf("unexpected Payment ID in published event: got %v want %v", event.ID, d.payment.ID)
			}

			if event.Status != d.newStatus {
				t.Errorf("unexpected Payment Status in published event: got %v want %v", event.Status, d.newStatus)
			}

			if event.ExternalID != d.payment.ExternalID {
				t.Errorf("unexpected ExternalID: got %v want %v", event.ExternalID, d.payment.ExternalID)
			}

			return response
		})
}
