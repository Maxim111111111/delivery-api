package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var ErrOrderNotFound = errors.New("order not found")

type Order struct {
	ID      int    `json:"id"`
	Address string `json:"address"`
	Price   int64  `json:"price"`
}

type Server struct {
	db *sql.DB
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DATABASE_URL")

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Success connection")

	srv := &Server{db: db}

	r := chi.NewRouter()

	r.Use(LoggingMiddleware)
	r.Use(RecoverMiddleware)

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)
		r.Get("/orders", srv.GetOrders)
		r.Get("/orders/{id}", srv.GetOrderByID)
		r.Delete("/orders/{id}", srv.DeleteOrder)
		r.Post("/orders", srv.CreateOrder)
		r.Put("/orders/{id}", srv.UpdateOrder)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
	log.Println("Server has running")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Останавливаю сервер...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Println("Сервер перестал ждать и был остановлен")
	} else {
		log.Println("Сервер успешно остановлен")
	}

}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("проверка прав для %s", r.URL.Path)
		// тут была бы проверка токена; пока пропускаем всех хах
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("паника: %v", rec)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) GetOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orders, err := getOrders(ctx, s.db)
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(orders)
	if err != nil {
		log.Printf("Error encoding orders: %v", err)
	}
}

func (s *Server) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	order, err := getOrderByID(ctx, s.db, intID)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Printf("GetOrderByID handler: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(order)
	if err != nil {
		log.Printf("Error encoding orders: %v", err)
	}
}

func (s *Server) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var o Order
	err := json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	order, err := createOrder(ctx, s.db, o)
	if err != nil {
		log.Printf("CreateOrder handler: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(order)
	if err != nil {
		log.Printf("Error encoding orders: %v", err)
	}
}

func (s *Server) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var o Order
	err = json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	order, err := updateOrder(ctx, s.db, intID, o)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Printf("UpdateOrder handler: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(order)
	if err != nil {
		log.Printf("Error encoding orders: %v", err)
	}
}

func (s *Server) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = deleteOrder(ctx, s.db, intID)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Printf("DeleteOrder handler: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getOrders(ctx context.Context, db *sql.DB) ([]Order, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, address, price FROM orders`)
	if err != nil {
		return nil, fmt.Errorf("getOrders query: %w", err)
	}
	defer rows.Close()

	orders := make([]Order, 0)
	for rows.Next() {
		var o Order
		err = rows.Scan(&o.ID, &o.Address, &o.Price)
		if err != nil {
			return nil, fmt.Errorf("getOrders scan: %w", err)
		}
		orders = append(orders, o)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("getOrders rows: %w", err)
	}
	return orders, nil
}

func getOrderByID(ctx context.Context, db *sql.DB, id int) (Order, error) {
	var o Order
	row := db.QueryRowContext(ctx, `SELECT id, address, price FROM orders WHERE id = $1`, id)
	err := row.Scan(&o.ID, &o.Address, &o.Price)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Order{}, ErrOrderNotFound
		}
		return Order{}, fmt.Errorf("getOrderByID: %w", err)
	}
	return o, nil
}

func createOrder(ctx context.Context, db *sql.DB, o Order) (Order, error) {
	row := db.QueryRowContext(ctx, `INSERT INTO orders (address, price) VALUES ($1, $2) RETURNING id`, o.Address, o.Price)
	err := row.Scan(&o.ID)
	if err != nil {
		return Order{}, fmt.Errorf("createOrder: %w", err)
	}
	return o, nil
}

func updateOrder(ctx context.Context, db *sql.DB, id int, o Order) (Order, error) {
	row := db.QueryRowContext(ctx, `UPDATE orders SET address = $1, price = $2 WHERE id = $3 RETURNING id, address, price`, o.Address, o.Price, id)
	err := row.Scan(&o.ID, &o.Address, &o.Price)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Order{}, ErrOrderNotFound
		}
		return Order{}, fmt.Errorf("updateOrder: %w", err)
	}
	return o, nil
}

func deleteOrder(ctx context.Context, db *sql.DB, id int) error {
	res, err := db.ExecContext(ctx, `DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleteOrder exec: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("deleteOrder rowsAffected: %w", err)
	}
	if rows == 0 {
		return ErrOrderNotFound
	}
	return nil
}
