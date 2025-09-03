package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"order_service/internal/domain"
	"order_service/internal/logger"
)

type Handler struct {
	service     domain.OrderService
	httpMetrics domain.HTTPMetrics
}

// NewHandler создает новый HTTP обработчик с внедренным сервисом заказов.
func NewHandler(service domain.OrderService, httpMetrics domain.HTTPMetrics) *Handler {
	logger.DebugLogger.Println("Initializing Handler")
	return &Handler{
		service:     service,
		httpMetrics: httpMetrics,
	}
}

// GetOrders возвращает HTTP обработчик для получения заказа по order_uid.
func (h *Handler) GetOrders() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer h.httpMetrics.ObserveRequest(start)
		h.httpMetrics.IncRequest()

		ctx := r.Context()
		orderUID := r.PathValue("order_uid")

		order, err := h.service.GetOrder(ctx, orderUID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			if errors.Is(err, domain.ErrOrderNotFound) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(ErrorResponse{Error: domain.ErrOrderNotFound.Error()}) //nolint:errcheck,gosec
			} else {
				logger.ErrorLogger.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{Error: domain.ErrInternalServer.Error()}) //nolint:errcheck,gosec
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(OrderResponse{Order: order}) //nolint:errcheck,gosec
	}
}
