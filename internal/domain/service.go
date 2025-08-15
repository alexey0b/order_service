package domain

import (
	"context"
)

type OrderService interface {
	GetOrder(ctx context.Context, orderUID string) (*Order, error)
	SaveOrder(ctx context.Context, order *Order) error
}
