package handler

import (
	"bytes"
	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPermissionMiddleware(t *testing.T) {
	permRepo := repo.NewInMemoryPermissionRepo()
	rolePermRepo := repo.NewInMemoryRolePermissionRepo()
	permRepo.CreatePermission(&model.Permission{Name: "manage_roles"})
	roleRepo := repo.NewInMemoryRoleRepo()
	roleRepo.CreateRole(&model.Role{ID: 1, Name: "admin"})
	rolePermRepo.AssignPermission(1, 1)

	r := gin.Default()
	r.POST("/protected",
		func(c *gin.Context) { c.Set("role_id", "1"); c.Next() }, // set role_id first
		PermissionMiddleware(rolePermRepo, permRepo, "manage_roles"),
		func(c *gin.Context) { c.String(200, "ok") },
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/protected", bytes.NewBufferString(""))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}
