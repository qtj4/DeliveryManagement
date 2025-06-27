package handler

import (
	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware writes an audit log for every protected action
func AuditMiddleware(audit repo.AuditLogRepository, action, resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := uint(0)
		if v, ok := c.Get("user_id"); ok {
			userID = v.(uint)
		}
		c.Next()
		success := c.Writer.Status() < 400
		log := &model.AuditLog{
			UserID:    userID,
			Action:    action,
			Resource:  resource,
			Success:   success,
			Timestamp: time.Now().Unix(),
		}
		audit.CreateAudit(log)
	}
}
