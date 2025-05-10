package repository

import (
	"context"
	"github.com/Buricz9/notification-service/internal/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, n domain.Notification) error
	FindByID(ctx context.Context, id uint) (domain.Notification, error)
	List(ctx context.Context) ([]domain.Notification, error)
	Due(ctx context.Context, nowUnix int64) ([]domain.Notification, error)
	Save(ctx context.Context, n domain.Notification) error
	Delete(ctx context.Context, id uint) error
}
