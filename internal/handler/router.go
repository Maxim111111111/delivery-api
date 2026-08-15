package handler

import (
	"delivery-api/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(orderHandler *OrderHandler, healthHandler *HealthHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logging)
	r.Use(middleware.Recover)
	r.Get("/health/liveness", healthHandler.Liveness)
	r.Get("/health/readiness", healthHandler.Readiness)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth)
		r.Get("/orders", orderHandler.GetOrders)
		r.Get("/orders/{id}", orderHandler.GetOrderByID)
		r.Delete("/orders/{id}", orderHandler.DeleteOrder)
		r.Post("/orders", orderHandler.CreateOrder)
		r.Put("/orders/{id}", orderHandler.UpdateOrder)
	})
	return r
}
