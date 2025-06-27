package handler

import (
	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
	"deliverymanagement/pkg/rabbitmq"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type DamageReportHandler struct {
	DamageReports repo.DamageReportRepository
	Publisher     rabbitmq.Publisher
}

// POST /api/damage-report (multipart: delivery_id, type, description, photo)
func (h *DamageReportHandler) CreateDamageReport(c *gin.Context) {
	deliveryID, _ := strconv.Atoi(c.PostForm("delivery_id"))
	damageType := c.PostForm("type")
	desc := c.PostForm("description")
	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "photo required"})
		return
	}
	defer file.Close()
	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 5MB)"})
		return
	}
	if !isAllowedImage(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type"})
		return
	}
	filename := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), deliveryID, filepath.Ext(header.Filename))
	path := filepath.Join("uploads", filename)
	out, err := os.Create(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save file"})
		return
	}
	defer out.Close()
	size, _ := io.Copy(out, file)
	report := &model.DamageReport{
		DeliveryID:  uint(deliveryID),
		Type:        damageType,
		Description: desc,
		PhotoPath:   filename,
		PhotoSize:   size,
		PhotoMime:   header.Header.Get("Content-Type"),
		Timestamp:   time.Now(),
	}
	h.DamageReports.CreateDamageReport(report)
	if h.Publisher != nil {
		h.Publisher.Publish("email.queue", map[string]interface{}{
			"event":       "damage.reported",
			"delivery_id": deliveryID,
			"photo":       filename,
			"file_size":   size,
			"mime_type":   header.Header.Get("Content-Type"),
		})
	}
	c.JSON(http.StatusOK, report)
}

func isAllowedImage(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}
