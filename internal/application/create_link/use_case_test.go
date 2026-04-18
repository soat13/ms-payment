package create_link_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/payment/internal/application/create_link"
	"github.com/soat13/payment/internal/application/ports/out"
	"github.com/soat13/payment/internal/application/ports/out/mock"
	"github.com/soat13/payment/internal/domain"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type (
	deps struct {
		repository        *mock.MockRepository
		publisher         *mock.MockTopicPublisher
		paymentProvider   *mock.MockPaymentProvider
		useCase           *create_link.RequestPaymentLinkUseCase
		id                uuid.UUID
		payment           *domain.Payment
		link              domain.Link
		providerPaymentID domain.ProviderID
	}
)

func TestRequestPaymentLinkUseCaseExecute(t *testing.T) {
	t.Parallel()

	t.Run("should request payment link and persist payment as processing", func(t *testing.T) {
		ctx := context.Background()
		deps := getDeps(t)

		mockRepositoryGetByIDSuccess(t, ctx, deps)
		mockPaymentProviderRequestLinkSuccess(t, ctx, deps)
		mockRepositorySaveSuccess(t, ctx, deps)
		mockPublisher(t, ctx, deps, nil)

		payment, err := deps.useCase.Execute(ctx, create_link.CreateLinkInput{
			ID: deps.id,
		})

		require.NoError(t, err)
		require.NotNil(t, payment)
		require.Equal(t, deps.payment.ID, payment.ID)
		require.Equal(t, domain.StatusProcessing, payment.Status)
		require.Equal(t, deps.link, *payment.Link)
		require.Equal(t, domain.MercadoPagoProviderName, *payment.Provider)
		require.Equal(t, deps.providerPaymentID, *payment.ProviderPaymentID)
	})

	t.Run("should return error when repository get by id fails", func(t *testing.T) {
		ctx := context.Background()
		deps := getDeps(t)

		expectedErr := errors.New("payment not found")

		deps.repository.EXPECT().
			GetByID(ctx, deps.id).
			Return(nil, expectedErr)

		payment, err := deps.useCase.Execute(ctx, create_link.CreateLinkInput{
			ID: deps.id,
		})

		require.ErrorIs(t, err, expectedErr)
		require.Nil(t, payment)
	})

	t.Run("should return error when payment provider fails", func(t *testing.T) {
		ctx := context.Background()
		deps := getDeps(t)

		expectedErr := errors.New("error requesting payment link")

		mockRepositoryGetByIDSuccess(t, ctx, deps)

		deps.paymentProvider.EXPECT().
			CreateLink(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, input out.CreateLinkInput) (*out.CreateLinkOutput, error) {
				if input.PaymentID != deps.payment.ID {
					t.Errorf("unexpected PaymentID: got %v want %v", input.PaymentID, deps.payment.ID)
				}

				if input.Amount != deps.payment.Amount {
					t.Errorf("unexpected Amount: got %v want %v", input.Amount, deps.payment.Amount)
				}

				if input.Description != deps.payment.Description {
					t.Errorf("unexpected Description: got %v want %v", input.Description, deps.payment.Description)
				}

				return nil, expectedErr
			})

		payment, err := deps.useCase.Execute(ctx, create_link.CreateLinkInput{
			ID: deps.id,
		})

		require.ErrorIs(t, err, expectedErr)
		require.Nil(t, payment)
	})

	t.Run("should return error when repository save fails", func(t *testing.T) {
		ctx := context.Background()
		deps := getDeps(t)

		expectedErr := errors.New("error saving payment")

		mockRepositoryGetByIDSuccess(t, ctx, deps)
		mockPaymentProviderRequestLinkSuccess(t, ctx, deps)

		deps.repository.EXPECT().
			Save(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, payment *domain.Payment) error {
				if payment.Status != domain.StatusProcessing {
					t.Errorf("unexpected Status: got %v want %v", payment.Status, domain.StatusProcessing)
				}

				if *payment.Link != deps.link {
					t.Errorf("unexpected Link: got %v want %v", payment.Link, deps.link)
				}

				if *payment.Provider != domain.MercadoPagoProviderName {
					t.Errorf("unexpected ProviderName: got %v want %v", payment.Provider, domain.MercadoPagoProviderName)
				}

				if *payment.ProviderPaymentID != deps.providerPaymentID {
					t.Errorf("unexpected ProviderID: got %v want %v", *payment.ProviderPaymentID, deps.providerPaymentID)
				}

				return expectedErr
			})

		payment, err := deps.useCase.Execute(ctx, create_link.CreateLinkInput{
			ID: deps.id,
		})

		require.ErrorIs(t, err, expectedErr)
		require.Nil(t, payment)
	})

	t.Run("should return error when publisher fails", func(t *testing.T) {
		ctx := context.Background()
		deps := getDeps(t)

		expectedErr := errors.New("error publishing payment")

		mockRepositoryGetByIDSuccess(t, ctx, deps)
		mockPaymentProviderRequestLinkSuccess(t, ctx, deps)
		mockRepositorySaveSuccess(t, ctx, deps)
		mockPublisher(t, ctx, deps, expectedErr)

		payment, err := deps.useCase.Execute(ctx, create_link.CreateLinkInput{
			ID: deps.id,
		})

		require.ErrorIs(t, err, expectedErr)
		require.Nil(t, payment)
	})
}

func getDeps(t *testing.T) deps {
	t.Helper()

	repository := mock.NewMockRepository(gomock.NewController(t))
	publisher := mock.NewMockTopicPublisher(gomock.NewController(t))
	paymentProvider := mock.NewMockPaymentProvider(gomock.NewController(t))

	amount, _ := money.New(12345)
	link := domain.Link("https://mercadopago.com/checkout/123")
	providerID := domain.ProviderID("mp-link-123")

	payment := &domain.Payment{
		ID:          uuid.New(),
		Amount:      amount,
		Description: "payment description",
		Status:      domain.StatusPending,
	}

	return deps{
		repository:        repository,
		publisher:         publisher,
		paymentProvider:   paymentProvider,
		useCase:           create_link.NewRequestPaymentLinkUseCase(repository, publisher, paymentProvider),
		id:                payment.ID,
		payment:           payment,
		link:              link,
		providerPaymentID: providerID,
	}
}

func mockRepositoryGetByIDSuccess(t *testing.T, ctx context.Context, d deps) {
	t.Helper()

	d.repository.EXPECT().
		GetByID(ctx, d.id).
		Return(d.payment, nil)
}

func mockPaymentProviderRequestLinkSuccess(t *testing.T, ctx context.Context, d deps) {
	t.Helper()

	d.paymentProvider.EXPECT().
		CreateLink(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, input out.CreateLinkInput) (*out.CreateLinkOutput, error) {
			if input.PaymentID != d.payment.ID {
				t.Errorf("unexpected PaymentID: got %v want %v", input.PaymentID, d.payment.ID)
			}

			if input.Amount != d.payment.Amount {
				t.Errorf("unexpected Amount: got %v want %v", input.Amount, d.payment.Amount)
			}

			if input.Description != d.payment.Description {
				t.Errorf("unexpected Description: got %v want %v", input.Description, d.payment.Description)
			}

			return &out.CreateLinkOutput{
				Link:       d.link,
				Provider:   domain.MercadoPagoProviderName,
				ProviderID: d.providerPaymentID,
			}, nil
		})
}

func mockRepositorySaveSuccess(t *testing.T, ctx context.Context, d deps) {
	t.Helper()

	d.repository.EXPECT().
		Save(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, payment *domain.Payment) error {
			if payment.ID != d.payment.ID {
				t.Errorf("unexpected Payment ID: got %v want %v", payment.ID, d.payment.ID)
			}

			if payment.Status != domain.StatusProcessing {
				t.Errorf("unexpected Status: got %v want %v", payment.Status, domain.StatusProcessing)
			}

			if *payment.Link != d.link {
				t.Errorf("unexpected Link: got %v want %v", payment.Link, d.link)
			}

			if *payment.Provider != domain.MercadoPagoProviderName {
				t.Errorf("unexpected ProviderName: got %v want %v", payment.Provider, domain.MercadoPagoProviderName)
			}

			if *payment.ProviderPaymentID != d.providerPaymentID {
				t.Errorf("unexpected ProviderID: got %v want %v", *payment.ProviderPaymentID, d.providerPaymentID)
			}

			return nil
		})
}

func mockPublisher(t *testing.T, ctx context.Context, deps deps, response any) {
	t.Helper()

	deps.publisher.EXPECT().
		Publish(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, event domain.StatusChangedEvent) any {
			if event.ID != deps.payment.ID {
				t.Errorf("unexpected Payment ID in published event: got %v want %v", event.ID, deps.payment.ID)
			}

			if event.Status != domain.StatusProcessing {
				t.Errorf("unexpected Payment Status in published event: got %v want %v", event.Status, domain.StatusProcessing)
			}

			if event.ExternalID != deps.payment.ExternalID {
				t.Errorf("unexpected ExternalID: got %v want %v", event.ExternalID, deps.payment.ExternalID)
			}

			return response
		})

}
