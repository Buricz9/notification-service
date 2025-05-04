// cmd/scheduler/main.go
package main

import (
	"context"
	"github.com/streadway/amqp"
	"log"
	"time"

	"github.com/Buricz9/notification-service/internal/config"
	"github.com/Buricz9/notification-service/internal/db"
	"github.com/Buricz9/notification-service/internal/model"
	"github.com/Buricz9/notification-service/internal/mq"
	"github.com/Buricz9/notification-service/internal/repo"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config.Load: %v", err)
	}

	database := db.MustConnect(cfg.DatabaseURL)
	defer database.Close()

	conn := mq.MustConnect(cfg.RabbitMQURL)
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("conn.Channel: %v", err)
	}
	defer ch.Close()

	notificationRepo := repo.NewPostgresRepo(database, ch)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	log.Println("Scheduler started, ticking every 30s")

	for range ticker.C {
		ctx := context.Background()
		var pending []model.Notification
		query := `
            SELECT * FROM notifications
            WHERE status = 'pending'
              AND send_at <= now()
              AND date_part('hour', send_at AT TIME ZONE timezone) BETWEEN 8 AND 20
            ORDER BY priority DESC, send_at ASC
            LIMIT 10;`
		if err := database.SelectContext(ctx, &pending, query); err != nil {
			log.Printf("scheduler: fetch pending: %v", err)
			continue
		}

		for _, n := range pending {
			if err := ch.Publish("", "notifications", false, false,
				amqp.Publishing{ContentType: "text/plain", Body: []byte(n.ID.String())}); err != nil {
				log.Printf("scheduler: publish %s: %v", n.ID, err)
				continue
			}
			if err := notificationRepo.MarkScheduled(ctx, n.ID); err != nil {
				log.Printf("scheduler: mark scheduled %s: %v", n.ID, err)
			} else {
				log.Printf("Scheduler: scheduled %s (channel=%s)", n.ID, n.Channel)
			}
		}
	}
}
