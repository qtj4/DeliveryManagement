package handler

import (
	"deliverymanagement/internal/repo"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware checks if the user has the required permission
func PermissionMiddleware(rolePerms repo.RolePermissionRepository, perms repo.PermissionRepository, required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleIDVal, exists := c.Get("role_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no role"})
			return
		}
		roleID, _ := strconv.Atoi(roleIDVal.(string))
		permsList, _ := rolePerms.GetPermissions(uint(roleID))
		for _, p := range permsList {
			if p.Name == required {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}
