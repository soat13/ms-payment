package domain

import "errors"

var (
	ErrInvalidAmount = errors.New("amount must be greater than zero")

	ErrConcurrentModification = errors.New("concurrent modification detected")
	ErrPaymentAlreadyExists   = errors.New("payment already exists")
	ErrPaymentNotFound        = errors.New("payment not found")
)
