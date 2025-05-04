// internal/mq/mq.go

package mq

import (
	"log"
	"time"

	"github.com/streadway/amqp"
)

// MustConnect dłużej i bardziej wytrwale próbuje połączyć z RabbitMQ
func MustConnect(url string) *amqp.Connection {
	var conn *amqp.Connection
	var err error
	// spróbuj aż 30 razy co 2s (czyli ~60s)
	for i := 1; i <= 30; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			return conn
		}
		log.Printf("RabbitMQ not ready (%d/30), retrying in 2s: %v", i, err)
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("Could not connect to RabbitMQ after retries: %v", err)
	return nil
}
