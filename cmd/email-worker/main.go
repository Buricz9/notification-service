// cmd/email-worker/main.go
package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/Buricz9/notification-service/internal/config"
	"github.com/Buricz9/notification-service/internal/db"
	"github.com/Buricz9/notification-service/internal/mq"
	"github.com/Buricz9/notification-service/internal/repo"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

func main() {
	// Seed random number generator for simulated success/failure
	rand.Seed(time.Now().UnixNano())

	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config.Load: %v", err)
	}

	// 2. Connect to Postgres (with retry)
	database := db.MustConnect(cfg.DatabaseURL)
	defer database.Close()

	// 3. Connect to RabbitMQ (with retry)
	conn := mq.MustConnect(cfg.RabbitMQURL)
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("conn.Channel: %v", err)
	}
	defer ch.Close()

	// 4. Declare (or ensure) the "notifications" queue exists
	q, err := ch.QueueDeclare(
		"notifications", // name
		true,            // durable
		false,           // auto-delete
		false,           // exclusive
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Fatalf("QueueDeclare: %v", err)
	}

	// 5. Initialize repository
	notificationRepo := repo.NewPostgresRepo(database, ch)

	// 6. Start consuming messages (manual ack)
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Consume: %v", err)
	}

	log.Println("Email worker ready, awaiting messages")
	for d := range msgs {
		// Parse notification ID from message body
		id, err := uuid.Parse(string(d.Body))
		if err != nil {
			log.Printf("invalid UUID %q: %v", d.Body, err)
			d.Ack(false)
			continue
		}

		ctx := context.Background()
		n, err := notificationRepo.GetByID(ctx, id)
		if err != nil {
			log.Printf("GetByID %s: %v", id, err)
			d.Ack(false)
			continue
		}

		// Simulate send: 50% chance of success
		if rand.Float64() < 0.5 {
			// success path
			if err := notificationRepo.UpdateStatus(ctx, id, "sent", n.RetryCount, nil); err != nil {
				log.Printf("UpdateStatus sent %s: %v", id, err)
			} else {
				log.Printf("Email sent %s", id)
			}
			d.Ack(false)
		} else {
			// failure path
			errMsg := "simulated failure"
			if n.RetryCount < 3 {
				// schedule retry
				if err := notificationRepo.UpdateStatus(ctx, id, "pending", n.RetryCount+1, &errMsg); err != nil {
					log.Printf("UpdateStatus retry %s: %v", id, err)
				} else {
					// requeue for retry
					if err := ch.Publish(
						"",           // exchange
						q.Name,       // routing key
						false, false, // mandatory, immediate
						amqp.Publishing{
							ContentType: "text/plain",
							Body:        []byte(id.String()),
						},
					); err != nil {
						log.Printf("Publish retry %s: %v", id, err)
					} else {
						log.Printf("Email retry %s (try %d)", id, n.RetryCount+1)
					}
				}
				d.Ack(false)
			} else {
				// give up after 3 tries
				if err := notificationRepo.UpdateStatus(ctx, id, "failed", n.RetryCount+1, &errMsg); err != nil {
					log.Printf("UpdateStatus failed %s: %v", id, err)
				} else {
					log.Printf("Email failed %s after %d tries", id, n.RetryCount+1)
				}
				d.Ack(false)
			}
		}
	}
}
