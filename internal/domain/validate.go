package domain

// ValidateOrder проверяет корректность данных заказа
func ValidateOrder(order *Order) error {
	if order.OrderUID == "" {
		return ErrOrderUIDRequired
	}
	if order.CustomerID == "" {
		return ErrCustomerIDRequired
	}
	if order.Payment.Amount <= 0 {
		return ErrInvalidPaymentAmount
	}
	if len(order.Items) == 0 {
		return ErrNoItems
	}
	if order.TrackNumber == "" {
		return ErrTrackNumberRequired
	}
	if order.Payment.Transaction == "" {
		return ErrTransactionRequired
	}

	for _, item := range order.Items {
		if item.ChrtID <= 0 {
			return ErrInvalidItemID
		}
		if item.Price <= 0 {
			return ErrInvalidItemPrice
		}
	}

	return nil
}
