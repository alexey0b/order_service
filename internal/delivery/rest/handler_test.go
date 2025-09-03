package rest_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"order_service/config"
	"order_service/internal/delivery/rest"
	"order_service/internal/domain"
	"order_service/internal/logger"
	"order_service/internal/mock"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const pattern = "/api/v1/order/{order_uid}"

type instance struct {
	inputOrderUID   string
	outputOrderData *domain.Order
	outputErr       error

	expectedStatusCode    int
	expectedOrderResponse rest.OrderResponse
	expectedErrorResponse rest.ErrorResponse
}

var (
	cfg = &config.Config{
		Serv: config.Server{
			Debug: false,
		},
	}

	validOrder *domain.Order = &domain.Order{
		OrderUID:          "b563feb7b2b84b6test",
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
			Transaction:  "b563feb7b2b84b6test",
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
				Rid:         "XXXXXXXXXXXXXXXXXXXXX",
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

	tbl []instance = []instance{
		// 1. Валидные входные данные и правильный ответ
		{
			inputOrderUID:   "b563feb7b2b84b6test",
			outputOrderData: validOrder,
			outputErr:       nil,

			expectedStatusCode:    http.StatusOK,
			expectedOrderResponse: rest.OrderResponse{Order: validOrder},
			expectedErrorResponse: rest.ErrorResponse{},
		},
		// 2. Несуществующий orderUID и ответ 404 NotFound
		{
			inputOrderUID:      "b563feb7b2b84b6testtt",
			outputOrderData:    nil,
			expectedStatusCode: http.StatusNotFound,
			outputErr:          domain.ErrOrderNotFound,

			expectedOrderResponse: rest.OrderResponse{},
			expectedErrorResponse: rest.ErrorResponse{Error: domain.ErrOrderNotFound.Error()},
		},
		// 3. Пустой orderUID и ответ 404 NotFound
		{
			inputOrderUID:   "",
			outputOrderData: nil,

			expectedStatusCode:    http.StatusNotFound,
			expectedOrderResponse: rest.OrderResponse{},
			expectedErrorResponse: rest.ErrorResponse{},
		},
		// 4. Ошибка в сервисе и ответ 500 InternalServerError
		{
			inputOrderUID:   "b563feb7b2b84b6error",
			outputOrderData: nil,

			expectedStatusCode:    http.StatusInternalServerError,
			outputErr:             errors.New("internal server error"),
			expectedErrorResponse: rest.ErrorResponse{Error: "internal server error"},
		},
	}
)

func TestGetOrder(t *testing.T) {
	logger.InitLogger(cfg)

	for i, testCase := range tbl {
		t.Run(fmt.Sprintf("test case №%d", i+1), func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockOrderService := mock.NewMockOrderService(ctrl)

			if testCase.inputOrderUID != "" {
				mockOrderService.
					EXPECT().
					GetOrder(gomock.Any(), testCase.inputOrderUID).
					Return(testCase.outputOrderData, testCase.outputErr)
			}

			mockHTTPpMetrics := mock.NewMockHTTPMetrics(ctrl)
			if testCase.inputOrderUID != "" {
				mockHTTPpMetrics.
					EXPECT().
					IncRequest()

				mockHTTPpMetrics.
					EXPECT().
					ObserveRequest(gomock.Any())
			}

			handler := rest.NewHandler(mockOrderService, mockHTTPpMetrics)
			mux := http.NewServeMux()
			mux.HandleFunc(pattern, handler.GetOrders())

			baseURL := "/api/v1/order"
			orderUID := testCase.inputOrderUID
			testURL, err := url.JoinPath(baseURL, orderUID)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, testURL, nil)
			respRec := httptest.NewRecorder()

			mux.ServeHTTP(respRec, req)

			var (
				actualOrderResponse rest.OrderResponse
				actualErrorResponse rest.ErrorResponse
			)

			require.Equal(t, testCase.expectedStatusCode, respRec.Code)

			if testCase.expectedOrderResponse.Order != nil {
				require.NoError(t, json.NewDecoder(respRec.Body).Decode(&actualOrderResponse))
				require.Equal(t, testCase.expectedOrderResponse, actualOrderResponse)
			} else if testCase.expectedErrorResponse.Error != "" {
				require.NoError(t, json.NewDecoder(respRec.Body).Decode(&actualErrorResponse))
				require.Equal(t, testCase.expectedErrorResponse, actualErrorResponse)
				return
			}
		})
	}
}
