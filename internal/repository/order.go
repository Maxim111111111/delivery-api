package repository

import (
	"context"
	"database/sql"
	"delivery-api/internal/apperror"
	"delivery-api/internal/model"
	"errors"
	"fmt"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) GetOrders(ctx context.Context) ([]model.Order, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, address, price, status, created_at FROM orders`)
	if err != nil {
		return nil, fmt.Errorf("GetOrders query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var o model.Order
		err = rows.Scan(&o.ID, &o.Address, &o.Price, &o.Status, &o.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("GetOrders scan: %w", err)
		}
		orders = append(orders, o)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetOrders rows: %w", err)
	}
	return orders, nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id int) (model.Order, error) {
	var o model.Order
	row := r.db.QueryRowContext(ctx, `SELECT id, address, price, status, created_at FROM orders WHERE id = $1`, id)
	err := row.Scan(&o.ID, &o.Address, &o.Price, &o.Status, &o.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Order{}, apperror.ErrOrderNotFound
		}
		return model.Order{}, fmt.Errorf("GetOrderByID scan: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `SELECT id, order_id, name, quantity, price FROM order_items WHERE order_id = $1`, o.ID)
	if err != nil {
		return model.Order{}, fmt.Errorf("GetOrderByID query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	items := make([]model.OrderItem, 0)
	for rows.Next() {
		var i model.OrderItem
		err = rows.Scan(&i.ID, &i.OrderID, &i.Name, &i.Quantity, &i.Price)
		if err != nil {
			return model.Order{}, fmt.Errorf("GetOrderByID items scan: %w", err)
		}
		items = append(items, i)
	}
	if err = rows.Err(); err != nil {
		return model.Order{}, fmt.Errorf("GetOrderByID rows:%w", err)
	}

	o.Items = items

	return o, nil
}

func (r *OrderRepository) CreateOrder(ctx context.Context, o model.Order) (model.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Order{}, fmt.Errorf("CreateOrder beginTx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	row := tx.QueryRowContext(ctx, `INSERT INTO orders (address, price) VALUES ($1, $2) RETURNING id, status, created_at`, o.Address, o.Price)
	err = row.Scan(&o.ID, &o.Status, &o.CreatedAt)
	if err != nil {
		return model.Order{}, fmt.Errorf("CreateOrder scan: %w", err)
	}

	for i := range o.Items {
		row := tx.QueryRowContext(ctx, `INSERT INTO order_items(order_id, name, quantity, price) VALUES ($1, $2, $3, $4) RETURNING id, order_id`, o.ID, o.Items[i].Name, o.Items[i].Quantity, o.Items[i].Price)
		err = row.Scan(&o.Items[i].ID, &o.Items[i].OrderID)
		if err != nil {
			return model.Order{}, fmt.Errorf("CreateOrder items scan: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return model.Order{}, fmt.Errorf("CreateOrder commit: %w", err)
	}

	return o, nil
}

func (r *OrderRepository) UpdateOrder(ctx context.Context, id int, o model.Order) (model.Order, error) {
	row := r.db.QueryRowContext(ctx, `UPDATE orders SET address = $1, price = $2 WHERE id = $3 RETURNING id, address, price, status, created_at`, o.Address, o.Price, id)
	err := row.Scan(&o.ID, &o.Address, &o.Price, &o.Status, &o.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Order{}, apperror.ErrOrderNotFound
		}
		return model.Order{}, fmt.Errorf("UpdateOrder scan: %w", err)
	}
	return o, nil
}

func (r *OrderRepository) DeleteOrder(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("DeleteOrder exec: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteOrder rowsAffected: %w", err)
	}
	if rows == 0 {
		return apperror.ErrOrderNotFound
	}
	return nil
}
