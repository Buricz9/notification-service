// cmd/api/main.go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Buricz9/notification-service/internal/config"
	"github.com/Buricz9/notification-service/internal/db"
	"github.com/Buricz9/notification-service/internal/mq"
	"github.com/Buricz9/notification-service/internal/repo"
	httpTransport "github.com/Buricz9/notification-service/internal/transport/http"
)

func main() {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config.Load: %v", err)
	}

	// 2. Connect to Postgres (retry)
	database := db.MustConnect(cfg.DatabaseURL)
	defer database.Close()

	// 3. Connect to RabbitMQ (wait aż broker wystartuje)
	conn := mq.MustConnect(cfg.RabbitMQURL)
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("conn.Channel: %v", err)
	}
	defer ch.Close()

	// 4. Repo + HTTP
	notificationRepo := repo.NewPostgresRepo(database, ch)
	handlers := httpTransport.NewHandlers(notificationRepo)
	router := httpTransport.NewRouter(handlers)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("API listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
