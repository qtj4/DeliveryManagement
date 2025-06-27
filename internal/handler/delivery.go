package handler

import (
	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
	"deliverymanagement/pkg/rabbitmq"
	"encoding/csv"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/xuri/excelize/v2"
)

type DeliveryHandler struct {
	Deliveries repo.DeliveryRepository
	Publisher  rabbitmq.Publisher
}

func (h *DeliveryHandler) CreateDelivery(c *gin.Context) {
	_, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		FromAddress string `json:"from_address"`
		ToAddress   string `json:"to_address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	delivery := &model.Delivery{
		FromAddress: req.FromAddress,
		ToAddress:   req.ToAddress,
		Status:      "CREATED",
	}
	if err := h.Deliveries.CreateDelivery(delivery); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Publish event to email.queue
	if h.Publisher != nil {
		h.Publisher.Publish("email.queue", map[string]interface{}{
			"event":       "delivery.created",
			"delivery_id": delivery.ID,
			"from":        delivery.FromAddress,
			"to":          delivery.ToAddress,
		})
	}
	c.JSON(http.StatusOK, delivery)
}

func (h *DeliveryHandler) GetDelivery(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	delivery, err := h.Deliveries.GetDelivery(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, delivery)
}

// JWT middleware
func JWTAuthMiddleware(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if len(header) < 8 || header[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}
		tokenStr := header[7:]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}
		c.Set("user_id", uint(claims["user_id"].(float64)))
		c.Set("role", claims["role"])
		c.Next()
	}
}

// Role middleware for dispatcher
func DispatcherOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := c.Get("role")
		if !ok || role != "dispatcher" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "dispatcher only"})
			return
		}
		c.Next()
	}
}

// Assign a courier to a delivery (dispatcher only)
func (h *DeliveryHandler) AssignDelivery(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		CourierID uint `json:"courier_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	delivery, err := h.Deliveries.GetDelivery(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	delivery.Status = "ASSIGNED"
	// Optionally, you could add delivery.CourierID = req.CourierID if model supports it
	c.JSON(http.StatusOK, delivery)
}

// Handler for scan events

type ScanEventHandler struct {
	ScanEvents repo.ScanEventRepository
}

// Middleware for courier or warehouse roles
func CourierOrWarehouseOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := c.Get("role")
		if !ok || (role != "courier" && role != "warehouse") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "courier or warehouse only"})
			return
		}
		c.Next()
	}
}

// POST /api/scan
func (h *ScanEventHandler) CreateScanEvent(c *gin.Context) {
	var req struct {
		DeliveryID uint   `json:"delivery_id"`
		EventType  string `json:"event_type"`
		Location   string `json:"location"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.EventType != "IN" && req.EventType != "OUT" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_type"})
		return
	}
	if req.Location == "" || req.DeliveryID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields"})
		return
	}
	event := &model.ScanEvent{
		DeliveryID: req.DeliveryID,
		EventType:  req.EventType,
		Location:   req.Location,
		Timestamp:  time.Now(),
	}
	if err := h.ScanEvents.CreateScanEvent(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Broadcast to WebSocket clients for this delivery
	hub.broadcast(fmt.Sprint(event.DeliveryID), map[string]interface{}{
		"event":       "scan.updated",
		"delivery_id": event.DeliveryID,
		"event_type":  event.EventType,
		"location":    event.Location,
		"timestamp":   event.Timestamp,
	})
	c.JSON(http.StatusOK, event)
}

func (h *DeliveryHandler) ExportDeliveries(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	deliveries, _ := h.Deliveries.ListDeliveries() // Assume this returns []model.Delivery
	if len(deliveries) <= 1000 {
		// Sync export
		if format == "csv" {
			c.Header("Content-Disposition", "attachment; filename=deliveries.csv")
			c.Header("Content-Type", "text/csv")
			w := csv.NewWriter(c.Writer)
			w.Write([]string{"ID", "FromAddress", "ToAddress", "Status"})
			for _, d := range deliveries {
				w.Write([]string{
					strconv.Itoa(int(d.ID)), d.FromAddress, d.ToAddress, d.Status,
				})
			}
			w.Flush()
			return
		} else if format == "xlsx" {
			f := excelize.NewFile()
			f.SetSheetRow("Sheet1", "A1", &[]string{"ID", "FromAddress", "ToAddress", "Status"})
			for i, d := range deliveries {
				row := []interface{}{d.ID, d.FromAddress, d.ToAddress, d.Status}
				f.SetSheetRow("Sheet1", fmt.Sprintf("A%d", i+2), &row)
			}
			c.Header("Content-Disposition", "attachment; filename=deliveries.xlsx")
			c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			f.Write(c.Writer)
			return
		}
		c.JSON(400, gin.H{"error": "invalid format"})
		return
	}
	// Async: publish job
	jobID := fmt.Sprintf("job-%d", rand.Int63())
	if h.Publisher != nil {
		h.Publisher.Publish("export.queue", map[string]interface{}{
			"job_id": jobID,
			"format": format,
			"email":  c.Query("email"),
		})
	}
	c.JSON(202, gin.H{"job_id": jobID})
}
