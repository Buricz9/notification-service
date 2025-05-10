package domain

// Priority defines the urgency of a notification
// (Low or High).
type Priority string

const (
	PriorityLow  Priority = "Low"
	PriorityHigh Priority = "High"
)

// Status defines the life-cycle state of a notification
// (Pending, Sent, Delivered, Failed, Cancelled).
type Status string

const (
	StatusPending   Status = "Pending"
	StatusSent      Status = "Sent"
	StatusDelivered Status = "Delivered"
	StatusFailed    Status = "Failed"
	StatusCancelled Status = "Cancelled"
)

// Notification is the core domain entity used across all application layers.
type Notification struct {
	ID          uint     `json:"id" gorm:"primaryKey"`
	Recipient   string   `json:"recipient"`
	Message     string   `json:"message"`
	CreatedAt   int64    `json:"createdAt"`
	ScheduledAt int64    `json:"scheduledAt"`
	Priority    Priority `json:"priority"`
	Status      Status   `json:"status"`
	Channel     string   `json:"channel"`
	TimeZone    string   `json:"timeZone"`
	RetryCnt    int      `json:"retryCnt"`
}
