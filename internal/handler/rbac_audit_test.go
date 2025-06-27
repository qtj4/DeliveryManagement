package handler

import (
	"bytes"
	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRoleCRUDAndAudit(t *testing.T) {
	roleRepo := repo.NewInMemoryRoleRepo()
	auditRepo := repo.NewInMemoryAuditLogRepo()
	h := &RBACHandler{Roles: roleRepo, Audit: auditRepo}
	r := gin.Default()
	r.POST("/roles", AuditMiddleware(auditRepo, "create_role", "role"), h.CreateRole)
	r.GET("/roles", h.ListRoles)

	// Create role
	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"name":"admin"}`)
	req, _ := http.NewRequest("POST", "/roles", body)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// List roles
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/roles", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)
	var roles []model.Role
	json.Unmarshal(w2.Body.Bytes(), &roles)
	assert.Equal(t, 1, len(roles))
	assert.Equal(t, "admin", roles[0].Name)

	// Audit log should have entry
	logs, _ := auditRepo.ListAuditLogs()
	assert.Equal(t, 1, len(logs))
	assert.Equal(t, "create_role", logs[0].Action)
	assert.Equal(t, "role", logs[0].Resource)
	assert.True(t, logs[0].Success)
}
