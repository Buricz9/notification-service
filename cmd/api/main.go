package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	httpdelivery "github.com/Buricz9/notification-service/internal/delivery/http"
	queuedelivery "github.com/Buricz9/notification-service/internal/delivery/queue"
	"github.com/Buricz9/notification-service/internal/domain"
	gormrepo "github.com/Buricz9/notification-service/internal/repository/gorm"
	"github.com/Buricz9/notification-service/internal/service"
)

func main() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("PGHOST"), os.Getenv("PGUSER"), os.Getenv("PGPASSWORD"), os.Getenv("PGDATABASE"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	if err := db.AutoMigrate(&domain.Notification{}); err != nil {
		log.Fatalf("AutoMigrate error: %v", err)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	publisher := queuedelivery.NewRedisPublisher(redisAddr)

	repo := gormrepo.NewGormNotificationRepository(db)
	svc := service.NewNotificationService(repo, publisher)

	router := gin.Default()
	httpdelivery.RegisterRoutes(router, svc)

	router.Run(":" + os.Getenv("PORT"))
}
