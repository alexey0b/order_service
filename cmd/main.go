package main

import (
	"context"
	"net/http"
	"order_service/config"
	"order_service/internal/delivery/rest"
	"order_service/internal/domain"
	"order_service/internal/infrastructure/cache"
	"order_service/internal/infrastructure/kafka/consumer"
	"order_service/internal/logger"
	"order_service/internal/request/repositoriy/postgres"
	"order_service/internal/usecase"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.ErrorLogger.Fatalln("Ошибка загрузки internal/config/config.yaml файла:", err)
	}

	logger.InitLogger(cfg)

	db, err := sqlx.ConnectContext(ctx, "pgx", config.GetDbConnString(cfg))
	if err != nil {
		logger.ErrorLogger.Fatalln("Не удалось подключиться к базе данных:", err)
	}
	defer db.Close()

	cache := cache.NewLRUCache(cfg)
	repo := postgres.NewRequestRepositoryPostgres(db)
	service := usecase.NewOrderRequestService(cache, repo)
	handler := rest.NewHandler(service)
	consumer := consumer.NewConsumer(cfg)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("ui")))
	mux.HandleFunc("GET /api/v1/order/{order_uid}", handler.GetOrders())

	serv := &http.Server{
		Addr:         config.GetServerAddr(cfg),
		Handler:      mux,
		ReadTimeout:  time.Duration(cfg.Serv.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Serv.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Serv.IdleTimeout) * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	if err := service.RestoreCache(ctx, cfg); err != nil {
		logger.ErrorLogger.Println("Error restoring cache:", err)
	}

	go func() {
		logger.InfoLogger.Println("Starting Kafka consumer...")

		ticker := time.NewTicker(time.Duration(cfg.Kafka.PollTimeout) * time.Millisecond)
		defer ticker.Stop()
		defer consumer.Close()

		for {
			select {
			case <-ctx.Done():
				logger.InfoLogger.Println("Kafka consumer is stopped")
				return
			case <-ticker.C:
				order, err := consumer.ReadMessage(ctx)
				if err != nil {
					logger.ErrorLogger.Println("Error reading message:", err)
					continue
				}

				if domain.ValidateOrder(order) != nil {
					logger.ErrorLogger.Println("Invalid order:", err)
					continue
				}

				if err := service.SaveOrder(ctx, order); err != nil {
					logger.ErrorLogger.Println("Error saving order:", err)
				}
			}
		}
	}()

	go func() {
		<-quit

		logger.InfoLogger.Println("Order Service is stopping...")

		timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Duration(cfg.Serv.ShutdowTimeout)*time.Second)
		defer timeoutCtxCancel()

		if err := serv.Shutdown(timeoutCtx); err != nil {
			logger.ErrorLogger.Fatalln("Order Service shutdown error:", err)
		}
		logger.InfoLogger.Println("Order Service is stopped")
	}()

	logger.InfoLogger.Println("Order Service is running...")

	if err := serv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.ErrorLogger.Fatalln("Order Service start error:", err)
	}
}
