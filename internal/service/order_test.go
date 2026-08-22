package service

import (
	"context"
	"delivery-api/internal/model"
	"testing"
)

var _ OrderRepository = (*fakeRepo)(nil)

const address = "г. Ижевск, ул. Героя России Ильфата-Закирова, д.2, кв.3"

type fakeRepo struct {
	order       model.Order
	err         error
	createCall  bool
	createModel model.Order
}

func (f *fakeRepo) GetOrders(ctx context.Context) ([]model.Order, error) {
	return nil, f.err
}

func (f *fakeRepo) GetOrderByID(ctx context.Context, id int) (model.Order, error) {
	return f.order, f.err
}

func (f *fakeRepo) CreateOrder(ctx context.Context, o model.Order) (model.Order, error) {
	f.createCall = true
	f.createModel = o
	return f.order, f.err
}

func (f *fakeRepo) UpdateOrder(ctx context.Context, id int, o model.Order) (model.Order, error) {
	return f.order, f.err
}

func (f *fakeRepo) DeleteOrder(ctx context.Context, id int) error {
	return f.err
}

func TestValidateOrder(t *testing.T) {
	tests := []struct {
		name    string
		order   model.Order
		wantErr bool
	}{
		{name: "correctOrder",
			order: model.Order{Address: address,
				Price: 137000,
				Items: []model.OrderItem{{
					Name:     "Пицца Маргарита",
					Quantity: 2,
					Price:    50000,
				}}},
			wantErr: false},
		{name: "emptyAddress",
			order: model.Order{Address: "",
				Price: 137000,
				Items: []model.OrderItem{{
					Name:     "Пицца Маргарита",
					Quantity: 2,
					Price:    50000,
				}}},
			wantErr: true},
		{name: "emptyItems",
			order: model.Order{Address: address,
				Price: 137000,
				Items: []model.OrderItem{}},
			wantErr: true},
		{name: "zeroItemsQuantity",
			order: model.Order{Address: address,
				Price: 137000,
				Items: []model.OrderItem{{
					Name:     "Пицца Маргарита",
					Quantity: 0,
					Price:    50000,
				}}},
			wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateOrder(test.order)
			if (err != nil) != test.wantErr {
				t.Errorf("validateOrder() error: %v, wantErr: %v", err, test.wantErr)
			}
		})
	}
}

func TestInvalidCreateOrder(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewOrderService(repo)
	_, err := svc.CreateOrder(context.Background(), model.Order{})
	if err == nil {
		t.Errorf("Ошибки нет")
	}
	if repo.createCall {
		t.Errorf("Репозиторий вызван при невалидных данных")
	}
}

func TestValidCreateOrder(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewOrderService(repo)
	_, err := svc.CreateOrder(context.Background(), model.Order{Address: address,
		Price: 137000,
		Items: []model.OrderItem{{
			Name:     "Пицца Маргарита",
			Quantity: 2,
			Price:    50000,
		}}})
	if err != nil {
		t.Errorf("Ошибка: %v", err)
	}
	if !repo.createCall {
		t.Errorf("Репозиторий не вызван при валидных данных")
	}
	if repo.createModel.Address != address {
		t.Errorf("Пришло:%q; Ожидалось:%q", repo.createModel.Address, address)
	}
}
