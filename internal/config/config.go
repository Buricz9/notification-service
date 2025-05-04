// internal/config/config.go
package config

import (
	"errors"
	"os"
)

// Config trzyma wszystkie ustawienia serwisu.
type Config struct {
	DatabaseURL string // np. postgres://notif:secret@postgres/notifications?sslmode=disable
	RabbitMQURL string // np. amqp://guest:guest@rabbitmq:5672/
	SMTPHost    string // tylko dla email-workera
	SMTPUser    string
	SMTPPass    string
}

// Load czyta zmienne środowiskowe i zwraca Config lub błąd, jeśli brak którejś z wymaganych.
func Load() (*Config, error) {
	c := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RabbitMQURL: os.Getenv("RABBITMQ_URL"),
		SMTPHost:    os.Getenv("SMTP_HOST"),
		SMTPUser:    os.Getenv("SMTP_USER"),
		SMTPPass:    os.Getenv("SMTP_PASS"),
	}

	// Walidacja wymaganych pól
	if c.DatabaseURL == "" {
		return nil, errors.New("missing env DATABASE_URL")
	}
	if c.RabbitMQURL == "" {
		return nil, errors.New("missing env RABBITMQ_URL")
	}
	// SMTP opcjonalne w API, lecz wymagane w email-workerze — można walidować tam
	return c, nil
}
