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
	"sync"
	"syscall"
	"time"
)

var mu sync.RWMutex

var ErrOrderNotFound = errors.New("order not found")

type Product struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Brand    string `json:"brand"`
	Color    string `json:"color"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}

type Order struct {
	ID      int    `json:"id"`
	Address string `json:"address"`
	Price   int64  `json:"price"`
}

type Server struct {
	db *sql.DB
}

var products = map[int]Product{
	1: {1, "Iphone 15 Pro", "Apple", "White", 87990, 78},
	2: {2, "Samsung A54", "Samsung", "Gray", 13789, 8},
	3: {3, "Samsung A05", "Samsung", "White Gold", 3770, 3},
	4: {4, "Vivo S30 Pro Mini", "Vivo", "Pink", 35390, 32},
	5: {5, "Vivo X300", "Vivo", "Black", 63890, 58},
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

	r.Get("/products", GetProducts)
	r.Get("/products/{id}", GetProductByID)
	r.Get("/slow", SlowHandler)

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)
		r.Post("/products", CreateProduct)
		r.Put("/products/{id}", UpdateProduct)
		r.Delete("/products/{id}", DeleteProduct)
		r.Get("/boom", BoomHandler)
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

func BoomHandler(w http.ResponseWriter, r *http.Request) {
	panic("бум")
}

func SlowHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Медленный хендлер в работе...")
	time.Sleep(time.Second * 15)
	_, err := w.Write([]byte("готово"))
	if err != nil {
		log.Printf("Ошибка записи ответа: %v", err)
	}
	log.Println("Медленный хендлер закончил работу.")
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

func GetProducts(w http.ResponseWriter, r *http.Request) {
	productsSlice := make([]Product, 0)

	mu.RLock()
	for _, product := range products {
		productsSlice = append(productsSlice, product)
	}
	mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(productsSlice)
	if err != nil {
		log.Printf("Error encoding product: %v", err)
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

func GetProductByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	mu.RLock()
	product, ok := products[intID]
	mu.RUnlock()
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(product)
	if err != nil {
		log.Printf("Error encoding product: %v", err)
	}
}

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	mu.Lock()
	products[product.ID] = product
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(product)
	if err != nil {
		log.Printf("Error encoding product: %v", err)
	}
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var product Product
	err = json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	product.ID = intID
	mu.Lock()
	_, ok := products[intID]
	if !ok {
		mu.Unlock()
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	products[intID] = product
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(product)
	if err != nil {
		log.Printf("Error encoding product: %v", err)
	}
}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	mu.Lock()
	_, ok := products[intID]
	if !ok {
		mu.Unlock()
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	delete(products, intID)
	mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
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
