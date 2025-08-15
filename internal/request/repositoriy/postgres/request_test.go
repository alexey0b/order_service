package postgres_test

import (
	"context"
	"fmt"
	"log"
	"order_service/config"
	"order_service/internal/domain"
	"order_service/internal/logger"
	"order_service/internal/request/repositoriy/postgres"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	psqlC  testcontainers.Container
	testDB *sqlx.DB
	repo   *postgres.RequestRepositoryPostgres
	cfg    *config.Config
)

var testOrders = []*domain.Order{
	{
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
	},
	{
		OrderUID:          "test_order_789",
		TrackNumber:       "WBILMTESTTRACK3",
		Entry:             "WBIL",
		Locale:            "ru",
		CustomerID:        "test2",
		InternalSignature: "",
		DeliveryService:   "dhl",
		ShardKey:          "10",
		SmID:              101,
		DateCreated:       "2024-01-08T10:15:00Z",
		OofShard:          "2",
		Delivery: domain.Delivery{
			Name:    "Ivan Ivanov",
			Phone:   "+79001234567",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Lenina St 1",
			Region:  "Moscow",
			Email:   "ivan@test.com",
		},
		Payment: domain.Payment{
			Transaction:  "test_order_789",
			Currency:     "RUB",
			Provider:     "sberpay",
			Amount:       2500,
			PaymentDt:    1234567891,
			Bank:         "sber",
			DeliveryCost: 500,
			GoodsTotal:   2000,
			CustomFee:    0,
		},
		Items: []domain.Item{
			{
				ChrtID:      9934932,
				TrackNumber: "WBILMTESTTRACK3",
				Price:       2000,
				Rid:         "ab4219087a764ae0btest789",
				Name:        "Nike Sneakers",
				Sale:        20,
				Size:        "42",
				TotalPrice:  1600,
				NmID:        2389217,
				Brand:       "Nike",
				Status:      200,
			},
		},
	},
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	cfg = &config.Config{Serv: config.Server{Debug: false}}
	logger.InitLogger(cfg)

	buildContext, err := filepath.Abs("./testdata")
	if err != nil {
		log.Fatalf("Failed to resolve absolute path: %v\n", err)
	}

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context: buildContext,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	psqlC, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}

	host, err := psqlC.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	mappedPort, err := psqlC.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatal(err)
	}

	psqlUrl := fmt.Sprintf("postgres://user:password@%s:%s/test_db?sslmode=disable", host, mappedPort.Port())

	testDB, err = sqlx.Connect("pgx", psqlUrl)
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}

	repo = postgres.NewRequestRepositoryPostgres(testDB)

	code := m.Run()

	testcontainers.TerminateContainer(psqlC)
	testDB.Close()
	os.Exit(code)
}

func TestSaveAndGetOrders(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanRepo(testDB) })

	t.Run("save_order_success", func(t *testing.T) {
		require.NoError(t, repo.SaveOrder(ctx, testOrders[0]))
	})

	t.Run("save_duplicate_order", func(t *testing.T) {
		// Повторное сохранение того же заказа должно пройти без ошибки (ON CONFLICT DO NOTHING)
		require.NoError(t, repo.SaveOrder(ctx, testOrders[0]))
	})

	t.Run("get_existing_order", func(t *testing.T) {
		order, err := repo.GetOrder(ctx, testOrders[0].OrderUID)
		require.NoError(t, err)
		require.NotNil(t, order)
		require.Equal(t, testOrders[0].OrderUID, order.OrderUID)
		require.Equal(t, testOrders[0].TrackNumber, order.TrackNumber)
		require.Equal(t, testOrders[0].Delivery.Name, order.Delivery.Name)
		require.Equal(t, testOrders[0].Payment.Transaction, order.Payment.Transaction)
		require.Equal(t, testOrders[0].Items, order.Items)
	})

	t.Run("get_nonexistent_order", func(t *testing.T) {
		order, err := repo.GetOrder(ctx, "nonexistent_order")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrOrderNotFound)
		require.Nil(t, order)
	})

	t.Run("get_orders", func(t *testing.T) {
		require.NoError(t, repo.SaveOrder(ctx, testOrders[1]))
		orders, err := repo.GetOrders(ctx, len(testOrders))
		require.NoError(t, err)
		require.Len(t, orders, len(testOrders))
		iExpected := len(testOrders) - 1
		for iActual := range orders {
			require.Equal(t, *testOrders[iExpected], *orders[iActual]) // Проверяем порядок: новые заказы первыми (DESC)
			iExpected--
		}
	})

	t.Run("get_orders_empty", func(t *testing.T) {
		cleanRepo(testDB)
		orders, err := repo.GetOrders(ctx, len(testOrders))
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrOrdersNotFound)
		require.Nil(t, orders)
	})
}

func cleanRepo(testDB *sqlx.DB) {
	query := `
    DELETE FROM delivery;
    DELETE FROM payment;
    DELETE FROM items;
    DELETE FROM orders;
	`

	_, err := testDB.Exec(query)
	if err != nil {
		log.Fatalf("Failed to clean repository: %v", err)
	}
}
