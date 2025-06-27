package handler

import (
	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type DeliveryHandler struct {
	Deliveries repo.DeliveryRepository
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
	c.JSON(http.StatusOK, event)
}

// Handler for damage reports

type DamageReportHandler struct {
	DamageReports repo.DamageReportRepository
}

// POST /api/damage-report
func (h *DamageReportHandler) CreateDamageReport(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
		return
	}
	deliveryIDStr := c.PostForm("delivery_id")
	typeStr := c.PostForm("type")
	desc := c.PostForm("description")
	if deliveryIDStr == "" || typeStr == "" || desc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields"})
		return
	}
	deliveryID, err := strconv.ParseUint(deliveryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid delivery_id"})
		return
	}
	var photoPath string
	file, header, err := c.Request.FormFile("photo")
	if err == nil && header != nil {
		defer file.Close()
		filename := filepath.Base(header.Filename)
		uniqueName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename)
		uploadDir := "uploads"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create uploads dir"})
			return
		}
		photoPath = filepath.Join(uploadDir, uniqueName)
		out, err := os.Create(photoPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save file"})
			return
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not write file"})
			return
		}
	}
	report := &model.DamageReport{
		DeliveryID:  uint(deliveryID),
		Type:        typeStr,
		Description: desc,
		PhotoPath:   photoPath,
		Timestamp:   time.Now(),
	}
	if err := h.DamageReports.CreateDamageReport(report); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}
