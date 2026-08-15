package handler

import (
	"context"
	"delivery-api/internal/apperror"
	"delivery-api/internal/model"
	"delivery-api/internal/service"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
	"time"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	orders, err := h.service.GetOrders(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			writeError(w, http.StatusServiceUnavailable, "service temporarily unavailable, try again later")
			return
		}
		log.Printf("GetOrders handler: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, orders)
}

func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	id := chi.URLParam(r, "id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.service.GetOrderByID(ctx, intID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			writeError(w, http.StatusServiceUnavailable, "service temporarily unavailable, try again later")
			return
		}
		if errors.Is(err, apperror.ErrOrderNotFound) {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		log.Printf("GetOrderByID handler: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	var o model.Order
	err := json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	order, err := h.service.CreateOrder(ctx, o)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			writeError(w, http.StatusServiceUnavailable, "service temporarily unavailable, try again later")
			return
		}
		log.Printf("CreateOrder handler: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	id := chi.URLParam(r, "id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	var o model.Order
	err = json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	order, err := h.service.UpdateOrder(ctx, intID, o)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			writeError(w, http.StatusServiceUnavailable, "service temporarily unavailable, try again later")
			return
		}
		if errors.Is(err, apperror.ErrOrderNotFound) {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		log.Printf("UpdateOrder handler: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	id := chi.URLParam(r, "id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	err = h.service.DeleteOrder(ctx, intID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			writeError(w, http.StatusServiceUnavailable, "service temporarily unavailable, try again later")
			return
		}
		if errors.Is(err, apperror.ErrOrderNotFound) {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		log.Printf("DeleteOrder handler: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
