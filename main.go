package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	dsn := "postgres://delivery_user:root@localhost:5433/delivery?sslmode=disable"

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
	orders, err := getOrders(s.db)
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
	id := chi.URLParam(r, "id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	order, err := getOrderByID(s.db, intID)
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

func getOrders(db *sql.DB) ([]Order, error) {
	rows, err := db.Query(`SELECT id, address, price FROM orders`)
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

func getOrderByID(db *sql.DB, id int) (Order, error) {
	var o Order
	row := db.QueryRow(`SELECT id, address, price FROM orders WHERE id = $1`, id)
	err := row.Scan(&o.ID, &o.Address, &o.Price)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Order{}, ErrOrderNotFound
		}
		return Order{}, fmt.Errorf("getOrderByID: %w", err)
	}
	return o, nil
}
