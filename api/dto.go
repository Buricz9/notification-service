package main

type CreateNotificationDTO struct {
	Recipient   string `json:"recipient"   binding:"required"`
	Message     string `json:"message"     binding:"required"`
	ScheduledAt int64  `json:"scheduledAt" binding:"required"`
	Priority    string `json:"priority"    binding:"required,oneof=Low High"`

	Channel  string `json:"channel"  binding:"required,oneof=push email"`
	TimeZone string `json:"timezone" binding:"required"` // IANA tz, np. "Europe/Warsaw"
}

type ModifyNotificationDTO struct {
	Recipient   *string `json:"recipient"`
	Message     *string `json:"message"`
	ScheduledAt *int64  `json:"scheduledAt"`
	Priority    *string `json:"priority"`

	Channel  *string `json:"channel"`
	TimeZone *string `json:"timezone"`
}
