package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
	"time"
)

// Model bazy 1-do-1 z C#
type Notification struct {
	ID          uint `gorm:"primaryKey"`
	Recipient   string
	Message     string
	CreatedAt   int64
	ScheduledAt int64
	Priority    string // Low | High
	Status      string // Pending | Sent | Delivered | Failed

	Channel  string // push | email
	TimeZone string // IANA
	RetryCnt int    // liczba wykonanych prób
}

// proste połączenie z PG – zmienne środowiskowe dostajemy z docker-compose
func initDB() *gorm.DB {
	dsn := fmt.Sprintf("host=db user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("PGUSER"), os.Getenv("PGPASSWORD"), os.Getenv("PGDATABASE"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("cannot connect db: %v", err)
	}
	// auto-migracja – tak jak EnsureCreated w EF
	if err := db.AutoMigrate(&Notification{}); err != nil {
		log.Fatalf("auto-migrate: %v", err)
	}
	return db
}

func connectRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{Addr: "redis:6379"})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("cannot connect redis: %v", err)
	}
	return rdb
}

func main() {
	db := initDB()
	router := gin.Default()

	rdb := connectRedis()
	defer rdb.Close()

	registerRoutes(router, db, rdb)
	_ = db

	router.Run(":8080")
}

func nowUnix() int64 {
	return time.Now().Unix()
}

func intToStr(i uint) string {
	return strconv.FormatUint(uint64(i), 10)
}
