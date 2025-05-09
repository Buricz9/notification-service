package main

type StatusDto struct {
	NotificationId int64  `json:"notificationId"`
	Status         string `json:"status"` // "Delivered" | "Failed"
	RetryCnt       int
}
