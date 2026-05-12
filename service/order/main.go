package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	authMiddleware := NewAuthMiddleware(envOrDefault("AUTH_VALIDATE_URL", "http://auth_service:8081/auth/validate"))
	repo := NewOrderRepository(databaseURL)

	redisDB, err := strconv.Atoi(envOrDefault("REDIS_DB", "0"))
	if err != nil {
		log.Fatalf("Invalid REDIS_DB: %v\n", err)
	}
	cacheStore, err := NewRedisOrderStore(
		envOrDefault("REDIS_ADDR", "redis:6379"),
		os.Getenv("REDIS_PASSWORD"),
		redisDB,
	)
	if err != nil {
		log.Fatalf("Unable to initialize order redis store: %v\n", err)
	}

	publisher, err := NewRabbitPublisher(envOrDefault("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"))
	if err != nil {
		log.Fatalf("Unable to initialize rabbit publisher: %v\n", err)
	}

	service := NewOrderService(repo, cacheStore, publisher)
	handler := NewOrderHandler(service)

	r := mux.NewRouter()
	r.Handle("/orders", authMiddleware.RequireRoles(handler.CreateOrderHandler, basicUserRoles()...)).Methods("POST")
	r.Handle("/orders", authMiddleware.RequireRoles(handler.ListOrdersHandler, basicUserRoles()...)).Methods("GET")
	r.Handle("/orders/{id}", authMiddleware.RequireRoles(handler.GetOrderHandler, basicUserRoles()...)).Methods("GET")
	r.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         ":8083",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Order Service is running on :8083")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
