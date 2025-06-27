package handler

import (
	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
	"testing"
)

func TestNotificationCreateAndList(t *testing.T) {
	repo := repo.NewInMemoryNotificationRepo()
	n := &model.Notification{UserID: 1, Type: "test", Message: "Hello"}
	repo.CreateNotification(n)
	list, _ := repo.ListNotifications(1)
	if len(list) != 1 || list[0].Message != "Hello" {
		t.Errorf("expected notification to be listed")
	}
}

func TestNotificationMarkRead(t *testing.T) {
	repo := repo.NewInMemoryNotificationRepo()
	n := &model.Notification{UserID: 1, Type: "test", Message: "Hello"}
	repo.CreateNotification(n)
	repo.MarkRead(n.ID)
	list, _ := repo.ListNotifications(1)
	if !list[0].IsRead {
		t.Errorf("expected notification to be marked as read")
	}
}
