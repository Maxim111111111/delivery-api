package main

import (
	"context"
	"database/sql"
	"delivery-api/internal/handler"
	"delivery-api/internal/repository"
	"delivery-api/internal/service"
	"errors"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

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

	defer func() { _ = db.Close() }()
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Success db connection")

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetConnMaxIdleTime(time.Minute * 2)

	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)
	orderHandler := handler.NewOrderHandler(orderService)
	healthHandler := handler.NewHealthHandler(db)

	r := chi.NewRouter()

	r.Use(LoggingMiddleware)
	r.Use(RecoverMiddleware)
	r.Get("/health/liveness", healthHandler.Liveness)
	r.Get("/health/readiness", healthHandler.Readiness)

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)
		r.Get("/orders", orderHandler.GetOrders)
		r.Get("/orders/{id}", orderHandler.GetOrderByID)
		r.Delete("/orders/{id}", orderHandler.DeleteOrder)
		r.Post("/orders", orderHandler.CreateOrder)
		r.Put("/orders/{id}", orderHandler.UpdateOrder)
	})

	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	log.Println("Starting server on :8080")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Println(err)
	case <-quit:
		log.Println("Останавливаю сервер...")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Println("Сервер перестал ждать и был остановлен")
		} else {
			log.Println("Сервер успешно остановлен")
		}
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
