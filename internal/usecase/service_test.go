package usecase_test

import (
	"context"
	"fmt"
	"testing"

	"order_service/config"
	"order_service/internal/domain"
	"order_service/internal/logger"
	"order_service/internal/mock"
	"order_service/internal/usecase"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type instance struct {
	inputOrderUID    string
	inputOrderData   *domain.Order
	outputOrderData  *domain.Order
	outputOrdersData []*domain.Order
	outputOk         bool
	outputErr        error

	expectedOrder *domain.Order
	expectedErr   error
}

var (
	cfg = &config.Config{
		Serv: config.Server{
			Debug: false,
		},
	}

	tblForGetOrder = []instance{
		// 1. В кеше есть заказ и вывод без ошибки
		{
			inputOrderUID:   "cached_order",
			outputOrderData: &domain.Order{OrderUID: "cached_order"},
			outputOk:        true,
			outputErr:       nil,

			expectedOrder: &domain.Order{OrderUID: "cached_order"},
			expectedErr:   nil,
		},
		// 2. В кеше нет заказа, получение через repo без ошибки
		{
			inputOrderUID:   "repo_order",
			inputOrderData:  &domain.Order{OrderUID: "repo_order"},
			outputOrderData: &domain.Order{OrderUID: "repo_order"},
			outputOk:        false,
			outputErr:       nil,

			expectedOrder: &domain.Order{OrderUID: "repo_order"},
			expectedErr:   nil,
		},
		// 3. В кеше нет заказа, получение через repo с ошибкой
		{
			inputOrderUID:   "error_order",
			outputOrderData: nil,
			outputOk:        false,
			outputErr:       domain.ErrOrderNotFound,

			expectedOrder: nil,
			expectedErr:   domain.ErrOrderNotFound,
		},
	}

	tblForSaveOrder = []instance{
		// 1. Сохранение заказа без ошибки
		{
			inputOrderData: &domain.Order{OrderUID: "save_order"},
			outputErr:      nil,

			expectedErr: nil,
		},
		// 2. Сохранение заказа с ошибкой
		{
			inputOrderData: &domain.Order{OrderUID: "save_order"},
			outputErr:      domain.ErrOrderNotFound,

			expectedErr: domain.ErrOrderNotFound,
		},
	}

	tblForRestoreOrder = []instance{
		// 1. Успешное восстановление
		{
			outputOrdersData: []*domain.Order{{OrderUID: "restore_order"}},
			outputErr:        nil,
			expectedErr:      nil,
		},
		// 2. Ошибка получения заказов
		{
			outputOrdersData: nil,
			outputErr:        domain.ErrOrderNotFound,
			expectedErr:      domain.ErrOrderNotFound,
		},
		// 3. Пустой список заказов
		{
			outputOrdersData: []*domain.Order{},
			outputErr:        nil,
			expectedErr:      nil,
		},
	}
)

func TestGetOrder(t *testing.T) {
	logger.InitLogger(cfg)

	for i, testCase := range tblForGetOrder {
		t.Run(fmt.Sprintf("test_'GetOrder'_case №%d", i+1), func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockOrderRepo := mock.NewMockOrderRepository(ctrl)
			mockOrderCache := mock.NewMockOrderCache(ctrl)
			service := usecase.NewOrderRequestService(mockOrderCache, mockOrderRepo)

			mockOrderCache.
				EXPECT().
				GetOrder(testCase.inputOrderUID).
				Return(testCase.outputOrderData, testCase.outputOk)

			if !testCase.outputOk {
				mockOrderRepo.
					EXPECT().
					GetOrder(gomock.Any(), testCase.inputOrderUID).
					Return(testCase.outputOrderData, testCase.outputErr)

				if testCase.expectedErr == nil {
					mockOrderCache.
						EXPECT().
						SaveOrder(testCase.inputOrderUID, testCase.inputOrderData).
						Return()
				}
			}

			order, err := service.GetOrder(context.TODO(), testCase.inputOrderUID)

			if testCase.expectedErr != nil {
				require.ErrorContains(t, err, testCase.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, testCase.expectedOrder, order)
		})
	}
}

func TestSaveOrder(t *testing.T) {
	logger.InitLogger(cfg)

	for i, testCase := range tblForSaveOrder {
		t.Run(fmt.Sprintf("test_'SaveOrder'_case №%d", i+1), func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockOrderRepo := mock.NewMockOrderRepository(ctrl)
			mockOrderCache := mock.NewMockOrderCache(ctrl)
			service := usecase.NewOrderRequestService(mockOrderCache, mockOrderRepo)

			mockOrderCache.
				EXPECT().
				SaveOrder(gomock.Any(), testCase.inputOrderData).
				Return()

			mockOrderRepo.
				EXPECT().
				SaveOrder(gomock.Any(), testCase.inputOrderData).
				Return(testCase.outputErr)

			err := service.SaveOrder(context.TODO(), testCase.inputOrderData)

			if testCase.expectedErr != nil {
				require.ErrorContains(t, err, testCase.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRestoreOrder(t *testing.T) {
	cfg.Capacity = 1
	logger.InitLogger(cfg)

	for i, testCase := range tblForRestoreOrder {
		t.Run(fmt.Sprintf("test_'RestoreOrder'_case №%d", i+1), func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockOrderRepo := mock.NewMockOrderRepository(ctrl)
			mockOrderCache := mock.NewMockOrderCache(ctrl)
			service := usecase.NewOrderRequestService(mockOrderCache, mockOrderRepo)

			mockOrderRepo.
				EXPECT().
				GetOrders(gomock.Any(), gomock.Any()).
				Return(testCase.outputOrdersData, testCase.outputErr)

			if testCase.expectedErr == nil && len(testCase.outputOrdersData) != 0 {
				mockOrderCache.
					EXPECT().
					SaveOrder(gomock.Any(), gomock.Any()).
					Return()
			}

			err := service.RestoreCache(context.TODO(), cfg)

			if testCase.expectedErr != nil {
				require.ErrorContains(t, err, testCase.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
