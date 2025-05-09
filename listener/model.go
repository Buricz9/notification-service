package main

type Notification struct {
	ID       uint `gorm:"primaryKey"`
	Status   string
	RetryCnt int
}
