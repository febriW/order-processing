package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	authMiddleware := NewAuthMiddleware(envOrDefault("AUTH_VALIDATE_URL", "http://auth_service:8081/auth/validate"))
	repo := NewProductRepository(databaseURL)
	service := NewProductService(repo)
	handler := NewProductHandler(service)

	r := mux.NewRouter()
	r.Handle("/products", authMiddleware.RequireRoles(adminRoles()...)(http.HandlerFunc(handler.CreateProductHandler))).Methods("POST")
	r.Handle("/products", authMiddleware.RequireRoles(basicUserRoles()...)(http.HandlerFunc(handler.ListProductsHandler))).Methods("GET")
	r.Handle("/products/{id}", authMiddleware.RequireRoles(basicUserRoles()...)(http.HandlerFunc(handler.GetProductHandler))).Methods("GET")
	r.Handle("/products/{id}", authMiddleware.RequireRoles(adminRoles()...)(http.HandlerFunc(handler.UpdateProductHandler))).Methods("PUT")
	r.Handle("/products/{id}", authMiddleware.RequireRoles(adminRoles()...)(http.HandlerFunc(handler.DeleteProductHandler))).Methods("DELETE")
	r.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         ":8082",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Product Service is running on :8082")
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
