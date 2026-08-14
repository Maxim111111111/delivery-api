package service

import (
	"context"
	"delivery-api/internal/model"
	"delivery-api/internal/repository"
)

type OrderService struct {
	repo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) GetOrders(ctx context.Context) ([]model.Order, error) {
	return s.repo.GetOrders(ctx)
}

func (s *OrderService) GetOrderByID(ctx context.Context, id int) (model.Order, error) {
	return s.repo.GetOrderByID(ctx, id)
}

func (s *OrderService) CreateOrder(ctx context.Context, o model.Order) (model.Order, error) {
	return s.repo.CreateOrder(ctx, o)
}

func (s *OrderService) UpdateOrder(ctx context.Context, id int, o model.Order) (model.Order, error) {
	return s.repo.UpdateOrder(ctx, id, o)
}

func (s *OrderService) DeleteOrder(ctx context.Context, id int) error {
	return s.repo.DeleteOrder(ctx, id)
}
