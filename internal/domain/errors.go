package domain

import "errors"

var (
	// Repository errors

	ErrOrderNotFound  = errors.New("order not found")
	ErrOrdersNotFound = errors.New("orders not found")

	// http errors

	ErrInternalServer = errors.New("internal server error")

	// Validation errors - business rules

	ErrOrderUIDRequired     = errors.New("order_uid is required")
	ErrCustomerIDRequired   = errors.New("customer_id is required")
	ErrTrackNumberRequired  = errors.New("track_number is required")
	ErrTransactionRequired  = errors.New("transaction is required")
	ErrInvalidPaymentAmount = errors.New("payment amount must be positive")
	ErrNoItems              = errors.New("order must have at least one item")
	ErrInvalidItemID        = errors.New("item chrt_id must be positive")
	ErrInvalidItemPrice     = errors.New("item price must be positive")
)
