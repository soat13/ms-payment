package domain_test

import (
	"testing"

	"github.com/soat13/payment/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestProviderName(t *testing.T) {
	t.Parallel()

	t.Run("valid provider name", func(t *testing.T) {
		provider := domain.ProviderName("mercado-pago")

		err := provider.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid provider name", func(t *testing.T) {
		provider := domain.ProviderName("")
		err := provider.Validate()
		assert.Error(t, err)
	})
}

func TestProviderID(t *testing.T) {
	t.Parallel()

	t.Run("valid provider id", func(t *testing.T) {
		provider := domain.ProviderID("valid_id")
		err := provider.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid provider id", func(t *testing.T) {
		provider := domain.ProviderID("")
		err := provider.Validate()
		assert.Error(t, err)
	})
}
