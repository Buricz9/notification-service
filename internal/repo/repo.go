// internal/repo/repo.go
package repo

import (
	"context"
	"time"

	"github.com/Buricz9/notification-service/internal/model"
	"github.com/google/uuid"
)

// NotificationRepository to interfejs operacji na powiadomieniach.
// Dzięki niemu łatwo podmienisz implementację (np. na mock w testach).
type NotificationRepository interface {
	// Create dodaje nowe powiadomienie i zwraca wygenerowane ID.
	Create(ctx context.Context, n *model.Notification) (uuid.UUID, error)

	// GetByID pobiera powiadomienie po jego UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*model.Notification, error)

	// UpdateStatus aktualizuje status, liczbę retry i ewentualny błąd.
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, retryCount int, errMsg *string) error

	ListPendingBefore(ctx context.Context, before time.Time) ([]model.Notification, error)
	MarkScheduled(ctx context.Context, id uuid.UUID) error
	// ListByStatus zwraca powiadomienia o podanym statusie
	ListByStatus(ctx context.Context, status string) ([]model.Notification, error)

	// Stats zwraca liczbę powiadomień pogrupowanych po statusie w podanym przedziale czasu
	Stats(ctx context.Context, from, to time.Time) (map[string]int, error)

	// ForceSend ustawia send_at = NOW() i zwraca, czy faktycznie zaktualizowano rekord
	ForceSend(ctx context.Context, id uuid.UUID) (bool, error)

	// Cancel zmienia status z pending na cancelled (lub failed) i zapobiega dalszym retry
	Cancel(ctx context.Context, id uuid.UUID) (bool, error)
}
