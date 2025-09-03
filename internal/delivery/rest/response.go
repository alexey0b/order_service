package rest

import "order_service/internal/domain"

type OrderResponse struct {
	Order *domain.Order `json:"order"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
