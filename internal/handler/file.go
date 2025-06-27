package handler

import (
	"deliverymanagement/internal/repo"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	DamageReports repo.DamageReportRepository
}

// GET /files/:filename
func (h *FileHandler) ServeFile(c *gin.Context) {
	filename := c.Param("filename")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	reports := h.DamageReports.ListDamageReports(0) // get all
	var allowed bool
	for _, r := range reports {
		if r.PhotoPath == filename {
			if role == "admin" || role == "dispatcher" || userID == r.DeliveryID {
				allowed = true
				break
			}
		}
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	path := filepath.Join("uploads", filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.File(path)
}
