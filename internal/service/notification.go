package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Buricz9/notification-service/internal/delivery/queue"
	"github.com/Buricz9/notification-service/internal/domain"
	"github.com/Buricz9/notification-service/internal/repository"
)

var (
	// ErrNotFound is returned when a notification cannot be found
	ErrNotFound = errors.New("notification not found")
	// ErrFinalized is returned when operation on a completed notification is attempted
	ErrFinalized = errors.New("notification already finalized")
)

// NotificationService encapsulates business logic for notifications.
type NotificationService struct {
	repo      repository.NotificationRepository
	publisher queue.Publisher
}

// NewNotificationService constructs a new service with given repo and queue publisher.
func NewNotificationService(r repository.NotificationRepository, p queue.Publisher) *NotificationService {
	return &NotificationService{repo: r, publisher: p}
}

// Create creates and persists a new notification.
func (s *NotificationService) Create(ctx context.Context, n domain.Notification) (domain.Notification, error) {
	n.CreatedAt = time.Now().Unix()
	n.Status = domain.StatusPending
	n.RetryCnt = 0
	if err := s.repo.Create(ctx, n); err != nil {
		return domain.Notification{}, err
	}
	return n, nil
}

// Modify updates mutable fields of an existing notification.
func (s *NotificationService) Modify(ctx context.Context, n domain.Notification) error {
	existing, err := s.repo.FindByID(ctx, n.ID)
	if err != nil {
		return ErrNotFound
	}
	if existing.Status != domain.StatusPending {
		return ErrFinalized
	}
	existing.Recipient = n.Recipient
	existing.Message = n.Message
	existing.ScheduledAt = n.ScheduledAt
	existing.Priority = n.Priority
	existing.Channel = n.Channel
	existing.TimeZone = n.TimeZone
	return s.repo.Save(ctx, existing)
}

// SendNow publishes the notification immediately to high-priority queue.
func (s *NotificationService) SendNow(ctx context.Context, id uint) error {
	n, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if n.Status != domain.StatusPending {
		return ErrFinalized
	}
	raw, _ := json.Marshal(n)
	if err := s.publisher.Publish(ctx, queue.TopicHigh, raw); err != nil {
		return err
	}

	n.Status = domain.StatusSent
	return s.repo.Save(ctx, n)
}

// Cancel sets the notification status to Cancelled if still pending.
func (s *NotificationService) Cancel(ctx context.Context, id uint) error {
	n, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if n.Status != domain.StatusPending {
		return ErrFinalized
	}
	n.Status = domain.StatusCancelled
	return s.repo.Save(ctx, n)
}

// EnqueueDue scans for due notifications and enqueues them respecting time windows.
func (s *NotificationService) EnqueueDue(ctx context.Context) error {
	now := time.Now().Unix()
	due, err := s.repo.Due(ctx, now)
	if err != nil {
		return err
	}

	for _, n := range due {
		// compute local time window
		loc, err := time.LoadLocation(n.TimeZone)
		if err != nil {
			loc = time.UTC
		}
		nowLoc := time.Now().In(loc)
		hour := nowLoc.Hour()
		if hour < 8 || hour >= 21 {
			continue
		}

		raw, _ := json.Marshal(n)
		topic := queue.TopicLow
		if n.Priority == domain.PriorityHigh {
			topic = queue.TopicHigh
		}
		if err := s.publisher.Publish(ctx, topic, raw); err != nil {
			continue
		}

		n.Status = domain.StatusSent
		s.repo.Save(ctx, n)
	}
	return nil
}

func (s *NotificationService) FindByID(ctx context.Context, id uint) (domain.Notification, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *NotificationService) List(ctx context.Context) ([]domain.Notification, error) {
	return s.repo.List(ctx)
}

func (s *NotificationService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
