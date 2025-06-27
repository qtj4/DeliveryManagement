package handler

import (
	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RBACHandler struct {
	Roles     repo.RoleRepository
	Perms     repo.PermissionRepository
	RolePerms repo.RolePermissionRepository
	Audit     repo.AuditLogRepository
}

func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req struct{ Name string }
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}
	role := &model.Role{Name: req.Name}
	h.Roles.CreateRole(role)
	c.JSON(http.StatusOK, role)
}

func (h *RBACHandler) ListRoles(c *gin.Context) {
	roles, _ := h.Roles.ListRoles()
	c.JSON(http.StatusOK, roles)
}

func (h *RBACHandler) DeleteRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.Roles.DeleteRole(uint(id))
	c.Status(http.StatusNoContent)
}

func (h *RBACHandler) CreatePermission(c *gin.Context) {
	var req struct{ Name string }
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}
	perm := &model.Permission{Name: req.Name}
	h.Perms.CreatePermission(perm)
	c.JSON(http.StatusOK, perm)
}

func (h *RBACHandler) ListPermissions(c *gin.Context) {
	perms, _ := h.Perms.ListPermissions()
	c.JSON(http.StatusOK, perms)
}

func (h *RBACHandler) AssignPermission(c *gin.Context) {
	var req struct{ RoleID, PermID uint }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}
	h.RolePerms.AssignPermission(req.RoleID, req.PermID)
	c.Status(http.StatusNoContent)
}

func (h *RBACHandler) ListAuditLogs(c *gin.Context) {
	logs, _ := h.Audit.ListAuditLogs()
	c.JSON(http.StatusOK, logs)
}
