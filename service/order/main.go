package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/febriW/order-processing/common/middleware"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	authMiddleware := middleware.NewAuthMiddleware(envOrDefault("AUTH_VALIDATE_URL", "http://auth_service:8081/auth/validate"), true)
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

	outboxInterval, err := time.ParseDuration(envOrDefault("OUTBOX_RELAY_INTERVAL", "2s"))
	if err != nil {
		log.Fatalf("Invalid OUTBOX_RELAY_INTERVAL: %v\n", err)
	}
	outboxClaimTTL, err := time.ParseDuration(envOrDefault("OUTBOX_CLAIM_TTL", "30s"))
	if err != nil {
		log.Fatalf("Invalid OUTBOX_CLAIM_TTL: %v\n", err)
	}
	outboxBaseBackoff, err := time.ParseDuration(envOrDefault("OUTBOX_BASE_BACKOFF", "2s"))
	if err != nil {
		log.Fatalf("Invalid OUTBOX_BASE_BACKOFF: %v\n", err)
	}
	outboxMaxBackoff, err := time.ParseDuration(envOrDefault("OUTBOX_MAX_BACKOFF", "2m"))
	if err != nil {
		log.Fatalf("Invalid OUTBOX_MAX_BACKOFF: %v\n", err)
	}
	outboxBatchSize, err := strconv.Atoi(envOrDefault("OUTBOX_BATCH_SIZE", "20"))
	if err != nil {
		log.Fatalf("Invalid OUTBOX_BATCH_SIZE: %v\n", err)
	}
	outboxMaxAttempts, err := strconv.Atoi(envOrDefault("OUTBOX_MAX_ATTEMPTS", "8"))
	if err != nil {
		log.Fatalf("Invalid OUTBOX_MAX_ATTEMPTS: %v\n", err)
	}

	service := NewOrderService(repo, cacheStore, publisher)
	handler := NewOrderHandler(service)
	relay := NewOutboxRelay(repo, publisher, outboxInterval, outboxBatchSize)
	relay.claimTTL = outboxClaimTTL
	relay.baseBackoff = outboxBaseBackoff
	relay.maxBackoff = outboxMaxBackoff
	relay.maxAttempts = outboxMaxAttempts
	go relay.Start()

	r := mux.NewRouter()
	r.Use(requestContextMiddleware("order_service"))
	r.Handle("/orders", authMiddleware.RequireRoles(middleware.BasicUserRoles()...)(http.HandlerFunc(handler.CreateOrderHandler))).Methods("POST")
	r.Handle("/orders", authMiddleware.RequireRoles(middleware.BasicUserRoles()...)(http.HandlerFunc(handler.ListOrdersHandler))).Methods("GET")
	r.Handle("/orders/{id}", authMiddleware.RequireRoles(middleware.BasicUserRoles()...)(http.HandlerFunc(handler.GetOrderHandler))).Methods("GET")
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
