package handler

import (
	"deliverymanagement/internal/model"
	"testing"
	"time"
)

func TestAnalyticsSummary(t *testing.T) {
	deliveries := &mockDeliveryRepo{
		deliveries: []*model.Delivery{
			{Status: "delivered", CreatedAt: mockTime(0), DeliveredAt: mockTime(2), CourierID: 1},
			{Status: "pending", CreatedAt: mockTime(0), CourierID: 2},
			{Status: "delivered", CreatedAt: mockTime(0), DeliveredAt: mockTime(4), CourierID: 1},
		},
	}
	h := &AnalyticsHandler{Deliveries: deliveries}
	// Call h.Summary with a test context and assert output (pseudo-code)
}

type mockDeliveryRepo struct {
	deliveries []*model.Delivery
}

func (m *mockDeliveryRepo) ListDeliveries() ([]*model.Delivery, error) { return m.deliveries, nil }

func mockTime(hours int) model.Time { return model.Time{}.Add(time.Duration(hours) * time.Hour) }
