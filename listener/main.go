package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// --------- stałe kanałów ----------
const channelStat = "notifications:status"

// --------- połączenia -------------
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

// --------- main -------------------
func main() {
	ctx := context.Background()
	db := dbConn()
	rdb := redisConn()

	sub := rdb.Subscribe(ctx, channelStat)
	log.Printf("listener ready – waiting on %s", channelStat)

	for msg := range sub.Channel() {
		var dto StatusDto
		if err := json.Unmarshal([]byte(msg.Payload), &dto); err != nil {
			log.Printf("bad payload: %v", err)
			continue
		}

		if err := db.Model(&Notification{}).
			Where("id = ?", dto.NotificationId).
			Updates(map[string]any{
				"status":    dto.Status,
				"retry_cnt": dto.RetryCnt,
			}).Error; err != nil {
			log.Printf("update err: %v", err)
		}

		log.Printf("status %d -> %s", dto.NotificationId, dto.Status)
	}
}
