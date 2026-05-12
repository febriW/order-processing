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
	redisDB, err := strconv.Atoi(envOrDefault("REDIS_DB", "0"))
	if err != nil {
		log.Fatalf("Invalid REDIS_DB: %v\n", err)
	}
	sessionStore, err := NewRedisSessionStore(
		envOrDefault("REDIS_ADDR", "redis:6379"),
		os.Getenv("REDIS_PASSWORD"),
		redisDB,
	)
	if err != nil {
		log.Fatalf("Unable to initialize session store: %v\n", err)
	}

	repo := NewAuthRepository(databaseURL)
	service := NewAuthService(repo, sessionStore)
	handler := NewAuthHandler(service)

	r := mux.NewRouter()
	r.HandleFunc("/auth/register", handler.RegisterHandler).Methods("POST")
	r.HandleFunc("/auth/login", handler.LoginHandler).Methods("POST")
	r.HandleFunc("/auth/refresh", handler.RefreshTokenHandler).Methods("POST")
	r.HandleFunc("/auth/validate", handler.ValidateTokenHandler).Methods("GET")
	r.HandleFunc("/auth/logout", handler.LogoutHandler).Methods("POST")
	r.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         ":8081",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Auth Service is running on :8081")
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
