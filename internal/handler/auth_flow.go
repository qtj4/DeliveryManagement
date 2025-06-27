package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"deliverymanagement/internal/repo"
	"deliverymanagement/pkg/rabbitmq"

	"github.com/gin-gonic/gin"
)

type AuthFlowHandler struct {
	Users     repo.UserRepository
	Publisher rabbitmq.Publisher
}

// POST /api/auth/reset-request
func (h *AuthFlowHandler) ResetRequest(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	user, err := h.Users.FindUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link will be sent."})
		return
	}
	token := generateToken()
	h.Users.SetResetToken(user.Email, token, time.Now().Add(24*time.Hour))
	resetURL := os.Getenv("BASE_URL") + "/api/auth/reset/" + token
	h.Publisher.Publish("email.queue", map[string]interface{}{
		"to":      user.Email,
		"subject": "Password Reset",
		"body":    "Reset your password: " + resetURL,
	})
	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link will be sent."})
}

// POST /api/auth/reset/:token
func (h *AuthFlowHandler) ResetPassword(c *gin.Context) {
	token := c.Param("token")
	var req struct {
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	users, _ := h.Users.ListUsers()
	for _, u := range users {
		if u.ResetToken == token && u.ResetTokenExpiry.After(time.Now()) {
			u.PasswordHash = req.NewPassword // hash in real impl
			u.ResetToken = ""
			u.ResetTokenExpiry = time.Time{}
			h.Users.UpdateUser(u)
			c.JSON(http.StatusOK, gin.H{"message": "Password updated"})
			return
		}
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
}

// POST /api/auth/verify/:token
func (h *AuthFlowHandler) VerifyEmail(c *gin.Context) {
	token := c.Param("token")
	users, _ := h.Users.ListUsers()
	for _, u := range users {
		if u.VerificationToken == token {
			u.IsVerified = true
			u.VerificationToken = ""
			h.Users.UpdateUser(u)
			c.JSON(http.StatusOK, gin.H{"message": "Email verified"})
			return
		}
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token"})
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
