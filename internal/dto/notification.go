package dto

type Create struct {
	Recipient   string `json:"recipient" binding:"required"`
	Message     string `json:"message" binding:"required"`
	ScheduledAt int64  `json:"scheduledAt" binding:"required"`
	Priority    string `json:"priority" binding:"required,oneof=Low High"`
	Channel     string `json:"channel" binding:"required,oneof=push email"`
	TimeZone    string `json:"timezone" binding:"required"`
}

type ModifyNotification struct {
	Recipient   *string `json:"recipient"`
	Message     *string `json:"message"`
	ScheduledAt *int64  `json:"scheduledAt"`
	Priority    *string `json:"priority"`
	Channel     *string `json:"channel"`
	TimeZone    *string `json:"timezone"`
}

type StatusDto struct {
	NotificationId uint   `json:"notificationId"`
	Status         string `json:"status"`
	RetryCnt       int    `json:"retryCnt"`
}
