package main

import "encoding/json"

type Notification struct {
	ID          int64  `json:"id"`
	Recipient   string `json:"recipient"`
	Message     string `json:"message"`
	ScheduledAt int64  `json:"scheduledAt"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
	RetryCnt    int    `json:"retryCnt"`
}

type StatusDto struct {
	NotificationId int64  `json:"notificationId"`
	Status         string `json:"status"`
	RetryCnt       int    `json:"retryCnt"`
}

func (n *Notification) Marshal() string {
	b, _ := json.Marshal(n)
	return string(b)
}

func (s *StatusDto) Marshal() string {
	b, _ := json.Marshal(s)
	return string(b)
}
