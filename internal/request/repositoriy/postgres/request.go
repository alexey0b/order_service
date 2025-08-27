package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"order_service/internal/domain"
	"order_service/internal/logger"

	"github.com/jmoiron/sqlx"
)

type RequestRepositoryPostgres struct {
	db *sqlx.DB
}

// NewRequestRepositoryPostgres создает новый PostgreSQL репозиторий с подключением к БД
func NewRequestRepositoryPostgres(db *sqlx.DB) *RequestRepositoryPostgres {
	logger.Debug("Initializing RequestRepositoryPostgres with database connection")
	return &RequestRepositoryPostgres{db: db}
}

const (
	// Запросы на вставку строк
	insertRowIntoOrders = `
	INSERT INTO orders
		(order_uid, track_number, entry, locale, internal_signature, 
		customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	ON CONFLICT (order_uid) DO NOTHING
	`

	insertRowIntoDelivery = `
	INSERT INTO delivery
		(order_uid, name, phone, zip, city, address, region, email)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (order_uid) DO NOTHING
	`

	insertRowIntoPayment = `
	INSERT INTO payment
		(order_uid, transaction, request_id, currency, provider, amount, 
		payment_dt, bank, delivery_cost, goods_total, custom_fee)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	ON CONFLICT (order_uid) DO NOTHING
	`

	insertRowIntoItems = `
	INSERT INTO items
		(order_uid, chrt_id, track_number, price, rid, 
		name, sale, size, total_price, nm_id, brand, status)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	ON CONFLICT (order_uid, chrt_id) DO NOTHING
	`

	// Запросы на получение строк
	getActualRowsFromOrders = `
	SELECT order_uid
	FROM orders
	ORDER BY date_created DESC
	LIMIT $1
	`

	getRowFromOrdersDeliveryAndPayment = `
	SELECT 
		o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
		o.customer_id, o.delivery_service, o.shardkey, o.sm_id, 
		o.date_created, o.oof_shard,
		d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
		p.transaction, p.request_id, p.currency, p.provider, p.amount,
		p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
	FROM orders o
		JOIN delivery d ON o.order_uid = d.order_uid
		JOIN payment p ON o.order_uid = p.order_uid
	WHERE o.order_uid = $1
	`

	getRowsFromOrdersDeliveryAndPayment = `
	SELECT 
		o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
		o.customer_id, o.delivery_service, o.shardkey, o.sm_id, 
		o.date_created, o.oof_shard,
		d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
		p.transaction, p.request_id, p.currency, p.provider, p.amount,
		p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
	FROM orders o
		JOIN delivery d ON o.order_uid = d.order_uid
		JOIN payment p ON o.order_uid = p.order_uid
	WHERE o.order_uid = ANY($1::text[])
	`

	getRowsFromItemsByOrderUID = `
	SELECT
		order_uid, chrt_id, track_number, price, rid,
		name, sale, size, total_price, nm_id, brand, status
	FROM items
	WHERE order_uid = $1
	`

	getRowsFromItemsByOrderUIDs = `
	SELECT
		order_uid, chrt_id, track_number, price, rid,
		name, sale, size, total_price, nm_id, brand, status
	FROM items
	WHERE order_uid = ANY($1::text[])
	`
)

// GetOrder получает всю информацию о заказе.
func (r *RequestRepositoryPostgres) GetOrder(ctx context.Context, orderUID string) (*domain.Order, error) {
	// Получаем строку из orders delivery и payment
	orderData := domain.OrderWithoutItems{}
	if err := r.db.GetContext(ctx, &orderData, getRowFromOrdersDeliveryAndPayment, orderUID); err != nil {
		return nil, fmt.Errorf("failed to select order row: %w", err)
	}

	itemsData := []domain.Item{}
	if err := r.db.SelectContext(ctx, &itemsData, getRowsFromItemsByOrderUID, orderUID); err != nil {
		return nil, fmt.Errorf("failed to select items rows: %w", err)
	}

	order, err := assembleOrder(orderUID, &orderData, itemsData)
	if err != nil {
		return nil, fmt.Errorf("failed to assemble order: %w", err)
	}

	return order, nil
}

// GetOrders получает список заказов с ограничением по количеству (сортировка по дате создания DESC).
func (r *RequestRepositoryPostgres) GetOrders(ctx context.Context, quantity int) ([]*domain.Order, error) {
	orderUIDs := []string{}

	if err := r.db.SelectContext(ctx, &orderUIDs, getActualRowsFromOrders, quantity); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrdersNotFound
		} else {
			return nil, fmt.Errorf("failed to select orders rows: %w", err)
		}
	}

	ordersData := []*domain.OrderWithoutItems{}
	if err := r.db.SelectContext(ctx, &ordersData, getRowsFromOrdersDeliveryAndPayment, orderUIDs); err != nil {
		return nil, fmt.Errorf("failed to select orders rows: %w", err)
	}

	itemsData := []domain.Item{}
	if err := r.db.SelectContext(ctx, &itemsData, getRowsFromItemsByOrderUIDs, orderUIDs); err != nil {
		return nil, fmt.Errorf("failed to select items rows: %w", err)
	}

	orders, err := assembleOrders(orderUIDs, ordersData, itemsData)
	if err != nil {
		return nil, fmt.Errorf("failed to assemble orders: %w", err)
	}

	return orders, nil
}

// SaveOrder сохраняет новый заказ.
func (r *RequestRepositoryPostgres) SaveOrder(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			logger.ErrorLogger.Println("failed to rollback transaction:", err)
		}
	}()

	// Вставляем строку в orders
	_, err = tx.ExecContext(ctx, insertRowIntoOrders,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return fmt.Errorf("failed to insert row into orders: %w", err)
	}

	// Вставляем строку в delivery
	_, err = tx.ExecContext(ctx, insertRowIntoDelivery,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City,
		order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("failed to insert row into delivery: %w", err)
	}

	// Вставляем строку в payment
	_, err = tx.ExecContext(ctx, insertRowIntoPayment,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("failed to insert row into payment: %w", err)
	}

	// Вставляем строки в items
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, insertRowIntoItems,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid,
			item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("failed to insert row into items: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// assembleOrder собирает заказ из структур *domain.OrderWithoutItems и []domain.Item в domain.Order
func assembleOrder(orderUID string, orderData *domain.OrderWithoutItems, itemsData []domain.Item) (*domain.Order, error) {
	for _, item := range itemsData {
		if item.OrderUID != orderUID {
			return nil, fmt.Errorf("mismatch orderUID with the orderUID in ordersData")
		}
	}

	order := &domain.Order{
		OrderUID:          orderData.OrderUID,
		TrackNumber:       orderData.TrackNumber,
		Entry:             orderData.Entry,
		Locale:            orderData.Locale,
		InternalSignature: orderData.InternalSignature,
		CustomerID:        orderData.CustomerID,
		DeliveryService:   orderData.DeliveryService,
		ShardKey:          orderData.ShardKey,
		SmID:              orderData.SmID,
		DateCreated:       orderData.DateCreated,
		OofShard:          orderData.OofShard,

		Delivery: domain.Delivery{
			Name:    orderData.Name,
			Phone:   orderData.Phone,
			Zip:     orderData.Zip,
			City:    orderData.City,
			Address: orderData.Address,
			Region:  orderData.Region,
			Email:   orderData.Email,
		},

		Payment: domain.Payment{
			Transaction:  orderData.Transaction,
			RequestID:    orderData.RequestID,
			Currency:     orderData.Currency,
			Provider:     orderData.Provider,
			Amount:       orderData.Amount,
			PaymentDt:    orderData.PaymentDt,
			Bank:         orderData.Bank,
			DeliveryCost: orderData.DeliveryCost,
			GoodsTotal:   orderData.GoodsTotal,
			CustomFee:    orderData.CustomFee,
		},

		Items: itemsData,
	}

	return order, nil
}

// assembleOrders собирает заказы из структур []*domain.OrderWithoutItems и []domain.Item в []*domain.Order
func assembleOrders(orderUIDs []string, ordersData []*domain.OrderWithoutItems, itemsData []domain.Item) ([]*domain.Order, error) {
	orderMap := make(map[string]*domain.Order)

	// Заполняем основные данные заказов
	for _, orderData := range ordersData {
		orderMap[orderData.OrderUID] = &domain.Order{
			OrderUID:          orderData.OrderUID,
			TrackNumber:       orderData.TrackNumber,
			Entry:             orderData.Entry,
			Locale:            orderData.Locale,
			InternalSignature: orderData.InternalSignature,
			CustomerID:        orderData.CustomerID,
			DeliveryService:   orderData.DeliveryService,
			ShardKey:          orderData.ShardKey,
			SmID:              orderData.SmID,
			DateCreated:       orderData.DateCreated,
			OofShard:          orderData.OofShard,

			Delivery: domain.Delivery{
				Name:    orderData.Name,
				Phone:   orderData.Phone,
				Zip:     orderData.Zip,
				City:    orderData.City,
				Address: orderData.Address,
				Region:  orderData.Region,
				Email:   orderData.Email,
			},

			Payment: domain.Payment{
				Transaction:  orderData.Transaction,
				RequestID:    orderData.RequestID,
				Currency:     orderData.Currency,
				Provider:     orderData.Provider,
				Amount:       orderData.Amount,
				PaymentDt:    orderData.PaymentDt,
				Bank:         orderData.Bank,
				DeliveryCost: orderData.DeliveryCost,
				GoodsTotal:   orderData.GoodsTotal,
				CustomFee:    orderData.CustomFee,
			},

			Items: make([]domain.Item, 0),
		}
	}

	// Добавляем items к соответствующим заказам
	for _, item := range itemsData {
		if order, exists := orderMap[item.OrderUID]; exists {
			order.Items = append(order.Items, item)
		}
	}

	fullOrders := make([]*domain.Order, len(orderUIDs))
	for i, uid := range orderUIDs {
		if order, exists := orderMap[uid]; exists {
			fullOrders[i] = order
		} else {
			return nil, fmt.Errorf("mismatch of orderUIDs array with the orderUID in ordersData")
		}
	}

	return fullOrders, nil
}
