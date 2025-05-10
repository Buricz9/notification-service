package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Buricz9/notification-service/internal/domain"
	"github.com/Buricz9/notification-service/internal/dto"
	gormrepo "github.com/Buricz9/notification-service/internal/repository/gorm"
)

func main() {
	dsn := "host=" + os.Getenv("PGHOST") +
		" user=" + os.Getenv("PGUSER") +
		" password=" + os.Getenv("PGPASSWORD") +
		" dbname=" + os.Getenv("PGDATABASE") +
		" port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB connection err: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis ping err: %v", err)
	}

	repo := gormrepo.NewGormNotificationRepository(db)
	subscriber := rdb.Subscribe(context.Background(), "notifications:status")
	ch := subscriber.Channel()

	log.Println("Listener started - awaiting status updates")

	for msg := range ch {
		var status dto.StatusDto
		if err := json.Unmarshal([]byte(msg.Payload), &status); err != nil {
			log.Printf("invalid status payload: %v", err)
			continue
		}

		n := domain.Notification{
			ID:       status.NotificationId,
			Status:   domain.Status(status.Status),
			RetryCnt: status.RetryCnt,
		}
		if err := repo.Save(context.Background(), n); err != nil {
			log.Printf("failed update: %v", err)
		}
		log.Printf("status %d -> %s", n.ID, n.Status)
	}
}
