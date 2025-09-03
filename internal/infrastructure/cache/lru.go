package cache

import (
	"time"

	"order_service/config"
	"order_service/internal/domain"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type LRUCache struct {
	cache *expirable.LRU[string, *domain.Order]
}

// NewLRUCache создает новый LRU кеш с TTL на основе конфигурации.
func NewLRUCache(cfg *config.Config) *LRUCache {
	var ttl time.Duration
	if cfg.Serv.Debug {
		ttl = time.Second * time.Duration(cfg.Ttl) // Debug: TTL in seconds
	} else {
		ttl = time.Hour * time.Duration(cfg.Ttl) // Production: TTL in hours
	}
	cache := expirable.NewLRU[string, *domain.Order](cfg.Capacity, nil, ttl)
	return &LRUCache{cache: cache}
}

// GetOrder получает заказ из кеша по order_uid.
func (c *LRUCache) GetOrder(orderUID string) (*domain.Order, bool) {
	order, ok := c.cache.Get(orderUID)
	return order, ok
}

// SaveOrder сохраняет заказ в кеш.
func (c *LRUCache) SaveOrder(orderUID string, order *domain.Order) {
	c.cache.Add(orderUID, order)
}
