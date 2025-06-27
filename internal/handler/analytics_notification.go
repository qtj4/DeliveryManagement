package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
	"deliverymanagement/pkg/ws"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	Deliveries repo.DeliveryRepository
	Users      repo.UserRepository
}

type NotificationHandler struct {
	Notifications repo.NotificationRepository
	WSHub         *ws.Hub
}

// GET /api/admin/analytics/summary
func (h *AnalyticsHandler) Summary(c *gin.Context) {
	deliveries, _ := h.Deliveries.ListDeliveries()
	total := len(deliveries)
	byStatus := map[string]int{}
	var totalTime float64
	var deliveredCount int
	for _, d := range deliveries {
		byStatus[d.Status]++
		if d.Status == "delivered" {
			deliveredCount++
			totalTime += d.DeliveredAt.Sub(d.CreatedAt).Hours()
		}
	}
	avgTime := 0.0
	if deliveredCount > 0 {
		avgTime = totalTime / float64(deliveredCount)
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "by_status": byStatus, "avg_time": avgTime})
}

// GET /api/admin/analytics/by-courier
func (h *AnalyticsHandler) ByCourier(c *gin.Context) {
	deliveries, _ := h.Deliveries.ListDeliveries()
	courierStats := map[uint]int{}
	for _, d := range deliveries {
		courierStats[d.CourierID]++
	}
	out := []gin.H{}
	for id, count := range courierStats {
		out = append(out, gin.H{"courier_id": id, "delivered": count})
	}
	c.JSON(http.StatusOK, out)
}

// GET /api/notifications
func (h *NotificationHandler) List(c *gin.Context) {
	userID := getUserIDFromContext(c)
	ns, _ := h.Notifications.ListNotifications(userID)
	c.JSON(http.StatusOK, ns)
}

// POST /api/notifications/:id/read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	h.Notifications.MarkRead(id)
	c.Status(http.StatusNoContent)
}

// PublishNotification publishes to repo, RabbitMQ, and WS
func (h *NotificationHandler) PublishNotification(n *model.Notification) {
	h.Notifications.CreateNotification(n)
	// TODO: publish to RabbitMQ
	if h.WSHub != nil {
		msg, _ := json.Marshal(n)
		h.WSHub.Publish("user:"+strconv.FormatUint(n.UserID, 10), msg)
	}
}

func getUserIDFromContext(c *gin.Context) uint64 {
	// TODO: extract from JWT/session
	return 1
}
