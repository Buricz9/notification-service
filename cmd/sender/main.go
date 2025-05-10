package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Buricz9/notification-service/internal/domain"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Buricz9/notification-service/internal/delivery/queue"
)

// Retry threshold
const maxRetries = 3

func main() {
	// Initialize Redis client
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Subscribe to both priority channels
	subscriber := rdb.Subscribe(ctx, queue.TopicHigh, queue.TopicLow)
	ch := subscriber.Channel()

	log.Println("Sender started, waiting for notifications...")
	rand.Seed(time.Now().UnixNano())

	handler := func(msg *redis.Message) {
		var n queue.Notification // import queue.Notification alias domain
		if err := json.Unmarshal([]byte(msg.Payload), &n); err != nil {
			log.Printf("invalid payload: %v", err)
			return
		}

		// Simulate send delay
		time.Sleep(500 * time.Millisecond)

		// Simulate 50% success
		success := rand.Float64() < 0.5
		if success {
			fmt.Printf("[OK] %s → %s\n", n.Recipient, n.Message)
			status := queue.StatusDto{NotificationId: n.ID, Status: string(domain.StatusDelivered), RetryCnt: n.RetryCnt}
			publishStatus(ctx, rdb, status)
			return
		}

		// Retry logic
		n.RetryCnt++
		if n.RetryCnt < maxRetries {
			// re-publish to high-priority
			raw, _ := json.Marshal(n)
			rdb.Publish(ctx, queue.TopicHigh, raw)
			fmt.Printf("[RETRY %d] %s\n", n.RetryCnt, n.Recipient)
			return
		}

		// Exhausted retries
		fmt.Printf("[FAIL 3/3] %s\n", n.Recipient)
		status := queue.StatusDto{NotificationId: n.ID, Status: string(domain.StatusFailed), RetryCnt: n.RetryCnt}
		publishStatus(ctx, rdb, status)
	}

	// Consume messages
	for {
		select {
		case msg := <-ch:
			handler(msg)
		case <-ctx.Done():
			log.Println("Sender shutting down")
			return
		}
	}
}
