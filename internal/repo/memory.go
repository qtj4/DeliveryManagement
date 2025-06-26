package repo

import (
	"deliverymanagement/internal/model"
	"errors"
	"sync"
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
