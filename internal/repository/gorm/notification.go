package gormrepo

import (
	"context"
	"github.com/Buricz9/notification-service/internal/domain"
	"gorm.io/gorm"
)

type GormNotificationRepository struct {
	db *gorm.DB
}

func NewGormNotificationRepository(db *gorm.DB) *GormNotificationRepository {
	return &GormNotificationRepository{db: db}
}

func (r *GormNotificationRepository) Create(ctx context.Context, n domain.Notification) error {
	return r.db.WithContext(ctx).Create(&n).Error
}

func (r *GormNotificationRepository) FindByID(ctx context.Context, id uint) (domain.Notification, error) {
	var n domain.Notification
	err := r.db.WithContext(ctx).First(&n, id).Error
	return n, err
}

func (r *GormNotificationRepository) List(ctx context.Context) ([]domain.Notification, error) {
	var list []domain.Notification
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *GormNotificationRepository) Due(ctx context.Context, nowUnix int64) ([]domain.Notification, error) {
	var due []domain.Notification
	err := r.db.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", domain.StatusPending, nowUnix).
		Find(&due).Error
	return due, err
}

func (r *GormNotificationRepository) Save(ctx context.Context, n domain.Notification) error {
	return r.db.WithContext(ctx).Save(&n).Error
}

func (r *GormNotificationRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Notification{}, id).Error
}
