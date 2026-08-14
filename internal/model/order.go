package model

import "time"

type Order struct {
	ID        int         `json:"id"`
	Address   string      `json:"address"`
	Price     int64       `json:"price"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	Items     []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID       int    `json:"id"`
	OrderID  int    `json:"order_id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    int64  `json:"price"`
}
