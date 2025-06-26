package repo

import (
	"deliverymanagement/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryUserRepo(t *testing.T) {
	repo := NewInMemoryUserRepo()
	user := &model.User{Email: "test@example.com", PasswordHash: "hash"}

	err := repo.CreateUser(user)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), user.ID)

	// Duplicate
	err = repo.CreateUser(&model.User{Email: "test@example.com", PasswordHash: "hash2"})
	assert.Error(t, err)

	// Find
	found, err := repo.FindUserByEmail("test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user, found)

	// Not found
	_, err = repo.FindUserByEmail("nope@example.com")
	assert.Error(t, err)
}

func TestInMemoryDeliveryRepo(t *testing.T) {
	repo := NewInMemoryDeliveryRepo()
	delivery := &model.Delivery{FromAddress: "A", ToAddress: "B", Status: "pending"}

	err := repo.CreateDelivery(delivery)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), delivery.ID)

	// Get
	found, err := repo.GetDelivery(delivery.ID)
	assert.NoError(t, err)
	assert.Equal(t, delivery, found)

	// Not found
	_, err = repo.GetDelivery(999)
	assert.Error(t, err)
}
