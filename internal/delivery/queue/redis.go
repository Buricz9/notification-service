// internal/delivery/queue/redis.go
package queue

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"

	queue "github.com/Buricz9/notification-service/internal/dto"
)

const (
	TopicHigh = "notifications:pending_high_priority"
	TopicLow  = "notifications:pending_low_priority"
	TopicStat = "notifications:status"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

type RedisPublisher struct {
	client *redis.Client
}

func NewRedisPublisher(addr string) *RedisPublisher {
	client := redis.NewClient(&redis.Options{Addr: addr})
	return &RedisPublisher{client: client}
}

func (r *RedisPublisher) Publish(ctx context.Context, topic string, payload []byte) error {
	return r.client.Publish(ctx, topic, payload).Err()
}

func PublishStatus(ctx context.Context, rdb *redis.Client, s queue.StatusDto) error {
	raw, _ := json.Marshal(s)
	return rdb.Publish(ctx, TopicStat, raw).Err()
}
