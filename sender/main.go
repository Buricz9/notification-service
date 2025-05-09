package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

// nazwy kanałów dokładnie jak w schedulerze
const (
	channelHigh = "notifications:pending_high_priority"
	channelLow  = "notifications:pending_low_priority"
	channelStat = "notifications:status"
)

func connectRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379", // nazwa kontenera z docker-compose
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("cannot connect redis: %v", err)
	}
	return rdb
}

func publishStatus(ctx context.Context, rdb *redis.Client, s StatusDto) error {
	return rdb.Publish(ctx, channelStat, s.Marshal()).Err()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	rdb := connectRedis()
	defer rdb.Close()

	handler := func(msg *redis.Message) {
		var n Notification
		if err := json.Unmarshal([]byte(msg.Payload), &n); err != nil {
			log.Printf("bad payload: %v", err)
			return
		}

		rand.Seed(time.Now().UnixNano())
		success := rand.Float64() < 0.5 // 50 %

		if success {
			// SUKCES → publish Delivered
			st := StatusDto{NotificationId: n.ID, Status: "Delivered", RetryCnt: n.RetryCnt}
			publishStatus(ctx, rdb, st)
			fmt.Printf("[OK] %s → %s\n", n.Recipient, n.Message)
			return
		}

		// NIEPOWODZENIE
		n.RetryCnt++
		if n.RetryCnt < 3 {
			// wracamy na HIGH
			raw, _ := json.Marshal(n)
			if err := rdb.Publish(ctx, channelHigh, raw).Err(); err != nil {
				log.Printf("requeue err: %v", err)
			} else {
				fmt.Printf("[RETRY %d] %s\n", n.RetryCnt, n.Recipient)
			}
			return
		}

		// trzecia porażka → Failed
		st := StatusDto{NotificationId: n.ID, Status: "Failed", RetryCnt: n.RetryCnt}
		publishStatus(ctx, rdb, st)
		fmt.Printf("[FAIL 3/3] %s\n", n.Recipient)
	}

	// subskrybuj obydwa kanały
	for _, ch := range []string{channelHigh, channelLow} {
		go func(c string) {
			sub := rdb.Subscribe(ctx, c)
			for m := range sub.Channel() {
				handler(m)
			}
		}(ch)
	}

	<-ctx.Done() // czekamy na Ctrl-C
}
