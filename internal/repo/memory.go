package repo

import (
	"deliverymanagement/internal/model"
	"errors"
	"sync"
	"time"
)

type InMemoryUserRepo struct {
	mu     sync.RWMutex
	users  map[string]*model.User // key: email
	nextID uint
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{
		users:  make(map[string]*model.User),
		nextID: 1,
	}
}

func (r *InMemoryUserRepo) CreateUser(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.Email]; exists {
		return errors.New("user already exists")
	}
	user.ID = r.nextID
	r.nextID++
	r.users[user.Email] = user
	return nil
}

func (r *InMemoryUserRepo) FindUserByEmail(email string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.users[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *InMemoryUserRepo) UpdateUser(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.Email]; !exists {
		return errors.New("user not found")
	}
	r.users[user.Email] = user
	return nil
}

func (r *InMemoryUserRepo) DeleteUser(email string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[email]; !exists {
		return errors.New("user not found")
	}
	delete(r.users, email)
	return nil
}

func (r *InMemoryUserRepo) ListUsers() ([]*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make([]*model.User, 0, len(r.users))
	for _, u := range r.users {
		users = append(users, u)
	}
	return users, nil
}

func (r *InMemoryUserRepo) SetVerificationToken(email, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, exists := r.users[email]
	if !exists {
		return errors.New("user not found")
	}
	user.VerificationToken = token
	return nil
}

func (r *InMemoryUserRepo) SetResetToken(email, token string, expiry time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, exists := r.users[email]
	if !exists {
		return errors.New("user not found")
	}
	user.ResetToken = token
	user.ResetTokenExpiry = expiry
	return nil
}

func (r *InMemoryUserRepo) VerifyUser(email string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, exists := r.users[email]
	if !exists {
		return errors.New("user not found")
	}
	user.IsVerified = true
	user.VerificationToken = ""
	return nil
}

type InMemoryDeliveryRepo struct {
	mu         sync.RWMutex
	deliveries map[uint]*model.Delivery
	nextID     uint
}

func NewInMemoryDeliveryRepo() *InMemoryDeliveryRepo {
	return &InMemoryDeliveryRepo{
		deliveries: make(map[uint]*model.Delivery),
		nextID:     1,
	}
}

func (r *InMemoryDeliveryRepo) CreateDelivery(delivery *model.Delivery) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delivery.ID = r.nextID
	r.nextID++
	r.deliveries[delivery.ID] = delivery
	return nil
}

func (r *InMemoryDeliveryRepo) GetDelivery(id uint) (*model.Delivery, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	delivery, exists := r.deliveries[id]
	if !exists {
		return nil, errors.New("delivery not found")
	}
	return delivery, nil
}

func (r *InMemoryDeliveryRepo) ListDeliveries() ([]model.Delivery, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]model.Delivery, 0, len(r.deliveries))
	for _, d := range r.deliveries {
		list = append(list, *d)
	}
	return list, nil
}

// In-memory scan event and tracking log support

type InMemoryScanEventRepo struct {
	mu       sync.RWMutex
	events   []*model.ScanEvent
	nextID   uint
	tracking map[uint][]*model.ScanEvent // delivery_id -> scan events (tracking log)
}

func NewInMemoryScanEventRepo() *InMemoryScanEventRepo {
	return &InMemoryScanEventRepo{
		events:   []*model.ScanEvent{},
		nextID:   1,
		tracking: make(map[uint][]*model.ScanEvent),
	}
}

func (r *InMemoryScanEventRepo) CreateScanEvent(event *model.ScanEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	event.ID = r.nextID
	r.nextID++
	r.events = append(r.events, event)
	r.tracking[event.DeliveryID] = append(r.tracking[event.DeliveryID], event)
	return nil
}

func (r *InMemoryScanEventRepo) ListScanEvents(deliveryID uint) []*model.ScanEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tracking[deliveryID]
}

// In-memory damage report support

type InMemoryDamageReportRepo struct {
	mu      sync.RWMutex
	reports []*model.DamageReport
	nextID  uint
}

func NewInMemoryDamageReportRepo() *InMemoryDamageReportRepo {
	return &InMemoryDamageReportRepo{
		reports: []*model.DamageReport{},
		nextID:  1,
	}
}

func (r *InMemoryDamageReportRepo) CreateDamageReport(report *model.DamageReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	report.ID = r.nextID
	r.nextID++
	r.reports = append(r.reports, report)
	return nil
}

func (r *InMemoryDamageReportRepo) ListDamageReports(deliveryID uint) []*model.DamageReport {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*model.DamageReport
	for _, rep := range r.reports {
		if rep.DeliveryID == deliveryID {
			result = append(result, rep)
		}
	}
	return result
}
