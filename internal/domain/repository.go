package domain

import (
	"context"
)

type OrderRepository interface {
	GetOrder(ctx context.Context, orderUID string) (*Order, error)
	GetOrders(ctx context.Context, quantity int) ([]*Order, error)
	SaveOrder(ctx context.Context, order *Order) error
}
