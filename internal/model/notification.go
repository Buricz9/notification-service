// internal/model/notification.go
package model

import (
	"time"

	"github.com/google/uuid"
)

// Notification odwzorowuje wiersz z tabeli notifications
type Notification struct {
	ID         uuid.UUID `db:"id"`
	UserID     uuid.UUID `db:"user_id"`
	Channel    string    `db:"channel"`
	Payload    []byte    `db:"payload"`
	SendAt     time.Time `db:"send_at"`
	Timezone   string    `db:"timezone"`
	Priority   int       `db:"priority"`
	Status     string    `db:"status"`
	RetryCount int       `db:"retry_count"`
	Error      *string   `db:"error"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
