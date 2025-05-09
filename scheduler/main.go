package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	channelHigh = "notifications:pending_high_priority"
	channelLow  = "notifications:pending_low_priority"
)

type Notification struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Recipient   string `json:"recipient"`
	Message     string `json:"message"`
	ScheduledAt int64  `json:"scheduledAt"`
	Priority    string `json:"priority"` // "Low" | "High"
	Status      string `json:"status"`   // "Pending" | "Sent" | ...

	Channel  string
	TimeZone string
	RetryCnt int
}

func dbConn() *gorm.DB {
	dsn := "host=db user=" + os.Getenv("PGUSER") +
		" password=" + os.Getenv("PGPASSWORD") +
		" dbname=" + os.Getenv("PGDATABASE") +
		" port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB connect err: %v", err)
	}
	return db
}

func redisConn() *redis.Client {
	rdb := redis.NewClient(&redis.Options{Addr: "redis:6379"})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis ping err: %v", err)
	}
	return rdb
}

func main() {
	db := dbConn()
	rdb := redisConn()
	ctx := context.Background()

	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		var due []Notification
		now := time.Now().Unix()

		if err := db.Where("status = ? AND scheduled_at <= ?", "Pending", now).
			Find(&due).Error; err != nil {
			log.Printf("query err: %v", err)
			continue
		}
		if len(due) == 0 {
			continue
		}

		for _, n := range due {
			raw, _ := json.Marshal(n)
			ch := channelLow
			if n.Priority == "High" {
				ch = channelHigh
			}
			if err := rdb.Publish(ctx, ch, raw).Err(); err != nil {
				log.Printf("publish err: %v", err)
			}
			// od razu oznaczamy jako „Sent”
			db.Model(&n).Update("status", "Sent")
		}
		log.Printf("published %d notifications", len(due))
	}
}
