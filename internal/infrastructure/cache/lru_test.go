package cache_test

import (
	"order_service/config"
	"order_service/internal/domain"
	"order_service/internal/infrastructure/cache"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var testOrder = &domain.Order{
	OrderUID:          "test_order_123",
	TrackNumber:       "WBILMTESTTRACK",
	Entry:             "WBIL",
	Locale:            "en",
	CustomerID:        "test",
	InternalSignature: "",
	DeliveryService:   "meest",
	ShardKey:          "9",
	SmID:              99,
	DateCreated:       "2024-01-07T06:22:08Z",
	OofShard:          "1",
	Delivery: domain.Delivery{
		Name:    "Test Testov",
		Phone:   "+9720000000",
		Zip:     "2639809",
		City:    "Kiryat Mozkin",
		Address: "Ploshad Mira 15",
		Region:  "Kraiot",
		Email:   "test@gmail.com",
	},
	Payment: domain.Payment{
		Transaction:  "test_order_123",
		Currency:     "USD",
		Provider:     "wbpay",
		Amount:       1817,
		PaymentDt:    1234567890,
		Bank:         "alpha",
		DeliveryCost: 1500,
		GoodsTotal:   317,
		CustomFee:    0,
	},
	Items: []domain.Item{
		{
			ChrtID:      9934930,
			TrackNumber: "WBILMTESTTRACK",
			Price:       453,
			Rid:         "ab4219087a764ae0btest",
			Name:        "Mascaras",
			Sale:        30,
			Size:        "0",
			TotalPrice:  317,
			NmID:        2389212,
			Brand:       "Vivienne Sabo",
			Status:      202,
		},
	},
}

func TestSaveAndGetOrder(t *testing.T) {
	cfg := &config.Config{
		Serv: config.Server{
			Debug: true,
		},
		Cache: config.Cache{
			Capacity: 2,
			Ttl:      15, // 15 sec
		},
	}

	t.Run("save_and_get_order", func(t *testing.T) {
        t.Parallel()

		lruCache := cache.NewLRUCache(cfg)

		lruCache.SaveOrder(testOrder.OrderUID, testOrder)

		order, ok := lruCache.GetOrder(testOrder.OrderUID)
		require.True(t, ok)
		require.Equal(t, testOrder, order)
	})

	t.Run("get_nonexistent_order", func(t *testing.T) {
        t.Parallel()

		lruCache := cache.NewLRUCache(cfg)

		order, ok := lruCache.GetOrder("nonexistent")
		require.False(t, ok)
		require.Nil(t, order)
	})

	t.Run("lru_eviction", func(t *testing.T) {
        t.Parallel()
        
		lruCache := cache.NewLRUCache(cfg)

		order1 := &domain.Order{OrderUID: "order1"}
		order2 := &domain.Order{OrderUID: "order2"}
		order3 := &domain.Order{OrderUID: "order3"}

		lruCache.SaveOrder("order1", order1)
		lruCache.SaveOrder("order2", order2)

		_, ok1 := lruCache.GetOrder("order1")
		_, ok2 := lruCache.GetOrder("order2")
		require.True(t, ok1)
		require.True(t, ok2)

		// Добавляем третий заказ (должен вытеснить первый)
		lruCache.SaveOrder("order3", order3)

		// Проверка вытеснения
		_, ok1 = lruCache.GetOrder("order1")
		_, ok2 = lruCache.GetOrder("order2")
		_, ok3 := lruCache.GetOrder("order3")

		require.False(t, ok1)
		require.True(t, ok2)
		require.True(t, ok3)
	})

	t.Run("ttl_expiration", func(t *testing.T) {
        t.Parallel()

		shortTtlCfg := &config.Config{
			Serv: config.Server{Debug: true},
			Cache: config.Cache{
				Capacity: 10,
				Ttl:      1, // 1 секунда вместо 15
			},
		}

		lruCache := cache.NewLRUCache(shortTtlCfg)

		lruCache.SaveOrder(testOrder.OrderUID, testOrder)

		order, ok := lruCache.GetOrder(testOrder.OrderUID)
		require.True(t, ok)
		require.Equal(t, testOrder, order)

		time.Sleep(1 * time.Second)

		// Проверка удаления заказа из кеша по истечении времени хранения
		_, ok = lruCache.GetOrder(testOrder.OrderUID)
		require.False(t, ok)
	})

	t.Run("update_existing_order", func(t *testing.T) {
        t.Parallel()

		lruCache := cache.NewLRUCache(cfg)

		lruCache.SaveOrder(testOrder.OrderUID, testOrder)

		// Обновляем заказ
		updatedOrder := &domain.Order{
			OrderUID:    testOrder.OrderUID,
			TrackNumber: "UPDATED_TRACK",
			Entry:       "UPDATED",
		}
		lruCache.SaveOrder(testOrder.OrderUID, updatedOrder)

		// Проверяем обновленный заказ
		order, ok := lruCache.GetOrder(testOrder.OrderUID)
		require.True(t, ok)
		require.Equal(t, updatedOrder, order)
		require.Equal(t, "UPDATED_TRACK", order.TrackNumber)
	})
}
