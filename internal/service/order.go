package service

import (
	"context"
	"delivery-api/internal/apperror"
	"delivery-api/internal/model"
	"fmt"
)

type OrderRepository interface {
	GetOrders(ctx context.Context) ([]model.Order, error)
	GetOrderByID(ctx context.Context, id int) (model.Order, error)
	CreateOrder(ctx context.Context, o model.Order) (model.Order, error)
	UpdateOrder(ctx context.Context, id int, o model.Order) (model.Order, error)
	DeleteOrder(ctx context.Context, id int) error
}

type OrderService struct {
	repo OrderRepository
}

func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func validateOrder(o model.Order) error {
	var messages []string

	if o.Address == "" {
		messages = append(messages, "address is required")
	}
	if o.Price < 0 {
		messages = append(messages, "price cannot be negative")
	}
	if len(o.Items) == 0 {
		messages = append(messages, "order must contain at least one item")
	}
	for i, item := range o.Items {
		if item.Name == "" {
			messages = append(messages, fmt.Sprintf("item[%d] name is required", i))
		}
		if item.Price < 0 {
			messages = append(messages, fmt.Sprintf("item[%d] price cannot be negative", i))
		}
		if item.Quantity <= 0 {
			messages = append(messages, fmt.Sprintf("item[%d] quantity must be greater than zero", i))
		}
	}
	if len(messages) == 0 {
		return nil
	}
	return &apperror.ValidationError{Messages: messages}
}

func (s *OrderService) GetOrders(ctx context.Context) ([]model.Order, error) {
	return s.repo.GetOrders(ctx)
}

func (s *OrderService) GetOrderByID(ctx context.Context, id int) (model.Order, error) {
	return s.repo.GetOrderByID(ctx, id)
}

func (s *OrderService) CreateOrder(ctx context.Context, o model.Order) (model.Order, error) {
	err := validateOrder(o)
	if err != nil {
		return model.Order{}, err
	}
	return s.repo.CreateOrder(ctx, o)
}

func (s *OrderService) UpdateOrder(ctx context.Context, id int, o model.Order) (model.Order, error) {
	err := validateOrder(o)
	if err != nil {
		return model.Order{}, err
	}
	return s.repo.UpdateOrder(ctx, id, o)
}

func (s *OrderService) DeleteOrder(ctx context.Context, id int) error {
	return s.repo.DeleteOrder(ctx, id)
}
