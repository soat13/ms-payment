package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/soat13/oficina-utils/pkg/money"
	"github.com/soat13/ms-payment/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPendingPayment(t *testing.T) {
	t.Parallel()

	t.Run("should return an error when amount is zero", func(t *testing.T) {
		amount, _ := money.New(0)
		payment, err := domain.NewPendingPayment(uuid.New(), amount, "some description")

		require.ErrorIs(t, err, domain.ErrInvalidAmount)
		require.Nil(t, payment)
	})

	t.Run("should create a new pending payment with valid input", func(t *testing.T) {
		externalID := uuid.New()
		amount, _ := money.New(123456)

		payment, err := domain.NewPendingPayment(externalID, amount, "some description")

		require.NoError(t, err)
		require.Equal(t, externalID, payment.ExternalID)
		require.Equal(t, amount.Cents, payment.Amount.Cents)
	})
}

func TestStatus(t *testing.T) {
	t.Parallel()
	t.Run("should return true if payment is pending", func(t *testing.T) {
		payment := &domain.Payment{
			Status: domain.StatusPending,
		}

		assert.True(t, payment.IsPending())
	})

	t.Run("should return false if payment is not pending", func(t *testing.T) {
		payment := &domain.Payment{
			Status: domain.StatusError,
		}

		assert.False(t, payment.IsPending())
	})

	t.Run("should return true if payment is processing", func(t *testing.T) {
		payment := &domain.Payment{
			Status: domain.StatusProcessing,
		}

		assert.True(t, payment.IsProcessing())
	})

	t.Run("should return false if payment is not processing", func(t *testing.T) {
		payment := &domain.Payment{
			Status: domain.StatusPending,
		}

		assert.False(t, payment.IsProcessing())
	})

	t.Run("should return true if payment failed", func(t *testing.T) {
		payment := &domain.Payment{
			Status: domain.StatusFailed,
		}

		assert.True(t, payment.IsFailed())
	})

	t.Run("should return false if payment did not fail", func(t *testing.T) {
		payment := &domain.Payment{
			Status: domain.StatusPending,
		}

		assert.False(t, payment.IsFailed())
	})
}

func TestPaymentStartProcessing(t *testing.T) {
	t.Parallel()

	t.Run("should return error when payment is not pending", func(t *testing.T) {
		payment := &domain.Payment{
			Status: domain.StatusProcessing,
		}

		link := domain.Link("https://example.com/payment")

		err := payment.StartProcessing(link, domain.MercadoPagoProviderName, "provider-payment-id")

		require.ErrorIs(t, err, domain.ErrInvalidStatusTransition)
	})

	t.Run("should return error when link is invalid", func(t *testing.T) {
		amount, _ := money.New(100)
		payment, err := domain.NewPendingPayment(uuid.New(), amount, "some description")
		require.NoError(t, err)

		invalidLink := domain.Link("invalid-link")

		err = payment.StartProcessing(invalidLink, domain.MercadoPagoProviderName, "provider-payment-id")

		require.Error(t, err)
	})

	t.Run("should return error when provider is invalid", func(t *testing.T) {
		amount, _ := money.New(100)
		payment, err := domain.NewPendingPayment(uuid.New(), amount, "some description")
		require.NoError(t, err)

		link := domain.Link("https://example.com/payment")
		require.NoError(t, err)

		invalidProvider := domain.ProviderName("invalid-provider")

		err = payment.StartProcessing(link, invalidProvider, "provider-payment-id")

		require.Error(t, err)
	})

	t.Run("should return error when provider payment id is invalid", func(t *testing.T) {
		amount, _ := money.New(100)
		payment, err := domain.NewPendingPayment(uuid.New(), amount, "some description")
		require.NoError(t, err)

		link := domain.Link("https://example.com/payment")
		require.NoError(t, err)

		err = payment.StartProcessing(link, domain.MercadoPagoProviderName, "")

		require.Error(t, err)
	})

	t.Run("should start processing payment with valid input", func(t *testing.T) {
		amount, _ := money.New(100)
		payment, err := domain.NewPendingPayment(uuid.New(), amount, "some description")
		require.NoError(t, err)

		link := domain.Link("https://example.com/payment")
		provider := domain.MercadoPagoProviderName
		providerID := domain.ProviderID("provider-payment-id")

		err = payment.StartProcessing(link, provider, providerID)

		require.NoError(t, err)
		require.Equal(t, domain.StatusProcessing, payment.Status)
		require.NotNil(t, payment.Link)
		require.Equal(t, link, *payment.Link)
		require.NotNil(t, payment.Provider)
		require.Equal(t, provider, *payment.Provider)
		require.NotNil(t, payment.ProviderPaymentID)
		require.Equal(t, providerID, *payment.ProviderPaymentID)
	})
}

func TestPaymentApplyAttemptResult(t *testing.T) {
	t.Parallel()

	t.Run("should return error when payment is not processing or failed", func(t *testing.T) {
		payment := &domain.Payment{
			Status: domain.StatusPending,
		}

		err := payment.ApplyAttemptResult(domain.StatusSucceeded)

		require.ErrorIs(t, err, domain.ErrInvalidStatusTransition)
	})

	t.Run("should apply succeeded status when payment is processing", func(t *testing.T) {
		payment := &domain.Payment{
			Status: domain.StatusProcessing,
		}

		err := payment.ApplyAttemptResult(domain.StatusSucceeded)

		require.NoError(t, err)
		require.Equal(t, domain.StatusSucceeded, payment.Status)
	})

	t.Run("should apply error status when payment is processing", func(t *testing.T) {
		payment := &domain.Payment{
			Status: domain.StatusProcessing,
		}

		err := payment.ApplyAttemptResult(domain.StatusError)

		require.NoError(t, err)
		require.Equal(t, domain.StatusError, payment.Status)
	})

	t.Run("should return error when new status is not an attempt result", func(t *testing.T) {
		payment := &domain.Payment{
			Status: domain.StatusProcessing,
		}

		err := payment.ApplyAttemptResult(domain.StatusPending)

		require.ErrorIs(t, err, domain.ErrInvalidStatusTransition)
	})

	t.Run("should return error when new status is processing", func(t *testing.T) {
		payment := &domain.Payment{
			Status: domain.StatusProcessing,
		}

		err := payment.ApplyAttemptResult(domain.StatusProcessing)

		require.ErrorIs(t, err, domain.ErrInvalidStatusTransition)
	})
}
