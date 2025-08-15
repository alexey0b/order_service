package usecase

import (
	"context"
	"fmt"
	"order_service/config"
	"order_service/internal/domain"
	"order_service/internal/logger"
)

type OrderRequestService struct {
	cache domain.OrderCache
	repo  domain.OrderRepository
}

// NewOrderRequestService создает новый сервис заказов с внедренными зависимостями кеша и репозитория
func NewOrderRequestService(cache domain.OrderCache, repo domain.OrderRepository) *OrderRequestService {
	logger.Debug("Initializing OrderRequestService")
	return &OrderRequestService{
		cache: cache,
		repo:  repo,
	}
}

// GetOrder получает заказ по order_uid с использованием cache
func (s *OrderRequestService) GetOrder(ctx context.Context, orderUID string) (*domain.Order, error) {
	var (
		order *domain.Order
		ok    bool
		err   error
	)

	if ctx.Err() != nil {
		return nil, fmt.Errorf("getting order cancelled: %w", ctx.Err())
	}

	order, ok = s.cache.GetOrder(orderUID)
	if !ok {
		order, err = s.repo.GetOrder(ctx, orderUID)
		if err != nil {
			return nil, fmt.Errorf("failed to get order: %w", err)
		}

		s.cache.SaveOrder(orderUID, order)
	}

	logger.InfoLogger.Printf("Successfully received order with orderUID: %s", order.OrderUID)

	return order, nil
}

// SaveOrder сохраняет заказ в кеш и репозиторий
func (s *OrderRequestService) SaveOrder(ctx context.Context, order *domain.Order) error {
	if ctx.Err() != nil {
		return fmt.Errorf("saving order cancelled: %w", ctx.Err())
	}

	s.cache.SaveOrder(order.OrderUID, order)

	if err := s.repo.SaveOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %v", err)
	}

	logger.InfoLogger.Printf("Successfully saved order with orderUID: %s", order.OrderUID)

	return nil
}

// RestoreCache восстанавливает кеш из БД при запуске приложения
func (s *OrderRequestService) RestoreCache(ctx context.Context, cfg *config.Config) error {
	logger.InfoLogger.Println("Restoring cache...")

	cap := cfg.Cache.Capacity
	orders, err := s.repo.GetOrders(ctx, cap)
	if err != nil {
		return fmt.Errorf("failed to get orders: %w", err)
	}

	if len(orders) == 0 {
		logger.InfoLogger.Println("No orders found for cache restoration")
		return nil
	}

	for _, order := range orders {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("cache restoration cancelled: %w", err)
		}
		s.cache.SaveOrder(order.OrderUID, order)
	}

	logger.InfoLogger.Println("Successfully restored cache")

	return nil
}
