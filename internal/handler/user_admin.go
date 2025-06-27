package handler

import (
	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserAdminHandler struct {
	Users repo.UserRepository
}

// GET /api/admin/users
func (h *UserAdminHandler) ListUsers(c *gin.Context) {
	users, _ := h.Users.ListUsers()
	c.JSON(http.StatusOK, users)
}

// POST /api/admin/users
func (h *UserAdminHandler) CreateUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required"`
		Role     string `json:"role" binding:"required,oneof=admin dispatcher reporter"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}
	// Hash password (simple example, replace with bcrypt in production)
	passwordHash := req.Password // TODO: use bcrypt
	user := &model.User{
		Email:        req.Email,
		Name:         req.Name,
		Role:         req.Role,
		PasswordHash: passwordHash,
	}
	if err := h.Users.CreateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// GET /api/admin/users/:id
func (h *UserAdminHandler) GetUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	users, _ := h.Users.ListUsers()
	for _, u := range users {
		if int(u.ID) == id {
			c.JSON(http.StatusOK, u)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

// PUT /api/admin/users/:id
func (h *UserAdminHandler) UpdateUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Email *string `json:"email" binding:"omitempty,email"`
		Name  *string `json:"name" binding:"omitempty"`
		Role  *string `json:"role" binding:"omitempty,oneof=admin dispatcher reporter"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationError(c, err)
		return
	}
	users, _ := h.Users.ListUsers()
	for _, u := range users {
		if int(u.ID) == id {
			if req.Email != nil {
				u.Email = *req.Email
			}
			if req.Name != nil {
				u.Name = *req.Name
			}
			if req.Role != nil {
				u.Role = *req.Role
			}
			if err := h.Users.UpdateUser(u); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, u)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

// DELETE /api/admin/users/:id
func (h *UserAdminHandler) DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	users, _ := h.Users.ListUsers()
	for _, u := range users {
		if int(u.ID) == id {
			h.Users.DeleteUser(u.Email)
			c.Status(http.StatusNoContent)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

// handleValidationError centralizes validation error formatting
func handleValidationError(c *gin.Context, err error) {
	if ve, ok := err.(validator.ValidationErrors); ok {
		errs := make(map[string]string)
		for _, fe := range ve {
			errs[fe.Field()] = fe.Tag()
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errs})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
