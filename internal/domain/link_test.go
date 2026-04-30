package domain_test

import (
	"testing"

	"github.com/soat13/ms-payment/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestLink(t *testing.T) {
	t.Run("valid link", func(t *testing.T) {
		link := "https://example.com/payment"

		err := domain.Link(link).Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid link", func(t *testing.T) {
		link := "httpexample.com/payment"
		err := domain.Link(link).Validate()
		assert.Error(t, err)
	})
}
