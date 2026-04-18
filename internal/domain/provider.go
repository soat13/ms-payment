package domain

import "errors"

type (
	ProviderName string
	ProviderID   string
)

const MercadoPagoProviderName ProviderName = "mercado-pago"

var (
	ErrInvalidProvider   = errors.New("invalid provider")
	ErrInvalidProviderID = errors.New("invalid provider ID")

	validProviders = map[ProviderName]struct{}{
		MercadoPagoProviderName: {},
	}
)

func (p ProviderName) Validate() error {
	if _, ok := validProviders[p]; !ok {
		return ErrInvalidProvider
	}
	return nil
}

func (p ProviderID) Validate() error {
	if len(p) == 0 {
		return ErrInvalidProviderID
	}

	return nil
}
