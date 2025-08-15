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
	getRowFromOrders = `
	SELECT
		order_uid, track_number, entry, locale, internal_signature,
		customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
	FROM orders
	WHERE order_uid = $1
	`

	getActualRowsFromOrders = `
	SELECT order_uid
	FROM orders
	ORDER BY date_created DESC
	LIMIT $1
	`

	getRowFromDelivery = `
	SELECT
		name, phone, zip, city, address, region, email
	FROM delivery
	WHERE order_uid = $1
	`

	getRowFromPayment = `
	SELECT
		transaction, request_id, currency, provider, amount,
		payment_dt, bank, delivery_cost, goods_total, custom_fee
	FROM payment
	WHERE order_uid = $1
	`

	getRowsFromItems = `
	SELECT
		chrt_id, track_number, price, rid,
		name, sale, size, total_price, nm_id, brand, status
	FROM items
	WHERE order_uid = $1
	`
)

// GetOrder получает всю информацию о заказе
func (r *RequestRepositoryPostgres) GetOrder(ctx context.Context, orderUID string) (*domain.Order, error) {
	order := domain.Order{}
	// Получаем строку из orders
	err := r.db.GetContext(ctx, &order, getRowFromOrders, orderUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		} else {
			return nil, fmt.Errorf("failed to select order's row: %w", err)
		}
	}

	// Получаем строку из delivery
	err = r.db.GetContext(ctx, &order.Delivery, getRowFromDelivery, orderUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		} else {
			return nil, fmt.Errorf("failed to select delivery's row: %w", err)
		}
	}

	// Получаем строку из payment
	err = r.db.GetContext(ctx, &order.Payment, getRowFromPayment, orderUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		} else {
			return nil, fmt.Errorf("failed to select payment's row: %w", err)
		}
	}

	// Получаем строки из items
	err = r.db.SelectContext(ctx, &order.Items, getRowsFromItems, orderUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		} else {
			return nil, fmt.Errorf("failed to select items's rows: %w", err)
		}
	}

	return &order, nil
}

// GetOrders получает список заказов с ограничением по количеству (сортировка по дате создания DESC)
func (r *RequestRepositoryPostgres) GetOrders(ctx context.Context, quantity int) ([]*domain.Order, error) {
	orderUIDs := []string{}
	orders := []*domain.Order{}

	// Получаем строки из orders
	err := r.db.SelectContext(ctx, &orderUIDs, getActualRowsFromOrders, quantity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrdersNotFound
		} else {
			return nil, fmt.Errorf("failed to select orders's rows: %w", err)
		}
	}

	for _, orderUID := range orderUIDs {
		order, err := r.GetOrder(ctx, orderUID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("failed to get orders: %w", err)
			} else {
				continue
			}
		}
		orders = append(orders, order)
	}

	if len(orders) == 0 {
		return nil, domain.ErrOrdersNotFound
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
