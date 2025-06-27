package repo

import (
	"deliverymanagement/internal/model"
	"sync"
)

type NotificationRepository interface {
	CreateNotification(n *model.Notification) error
	ListNotifications(userID uint64) ([]*model.Notification, error)
	MarkRead(id uint64) error
}

type InMemoryNotificationRepo struct {
	mu            sync.RWMutex
	notifications []*model.Notification
}

func NewInMemoryNotificationRepo() *InMemoryNotificationRepo {
	return &InMemoryNotificationRepo{}
}

func (r *InMemoryNotificationRepo) CreateNotification(n *model.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	n.ID = uint64(len(r.notifications) + 1)
	r.notifications = append(r.notifications, n)
	return nil
}

func (r *InMemoryNotificationRepo) ListNotifications(userID uint64) ([]*model.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*model.Notification
	for _, n := range r.notifications {
		if n.UserID == userID {
			out = append(out, n)
		}
	}
	return out, nil
}

func (r *InMemoryNotificationRepo) MarkRead(id uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, n := range r.notifications {
		if n.ID == id {
			n.IsRead = true
			return nil
		}
	}
	return nil
}
