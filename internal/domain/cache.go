package domain

type OrderCache interface {
	GetOrder(orderUID string) (*Order, bool)
	SaveOrder(orderUID string, order *Order)
}
