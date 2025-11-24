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
	repo := NewAuthRepository(databaseURL)
	service := NewAuthService(repo)
	handler := NewAuthHandler(service)

	r := mux.NewRouter()
	r.HandleFunc("/auth/register", handler.RegisterHandler).Methods("POST")
	r.HandleFunc("/auth/login", handler.LoginHandler).Methods("POST")
	r.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         ":8083",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Auth Service is running on :8083")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
