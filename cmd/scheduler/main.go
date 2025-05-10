package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	queuedelivery "github.com/Buricz9/notification-service/internal/delivery/queue"
	"github.com/Buricz9/notification-service/internal/domain"
	gormrepo "github.com/Buricz9/notification-service/internal/repository/gorm"
	"github.com/Buricz9/notification-service/internal/service"
)

func main() {
	// Load configuration from env
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("PGHOST"), os.Getenv("PGUSER"), os.Getenv("PGPASSWORD"), os.Getenv("PGDATABASE"))

	// Initialize GORM DB
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto-migrate domain model
	if err := db.AutoMigrate(&domain.Notification{}); err != nil {
		log.Fatalf("failed to auto-migrate: %v", err)
	}

	// Initialize Redis publisher
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	publisher := queuedelivery.NewRedisPublisher(redisAddr)

	// Wire repository and service
	repo := gormrepo.NewGormNotificationRepository(db)
	svc := service.NewNotificationService(repo, publisher)

	// Scheduler ticker every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	ctx := context.Background()
	log.Println("Scheduler started, enqueuing due notifications every 30s")

	for range ticker.C {
		if err := svc.EnqueueDue(ctx); err != nil {
			log.Printf("error enqueuing due notifications: %v", err)
		}
	}
}
