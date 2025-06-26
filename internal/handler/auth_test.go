package handler

import (
	"bytes"
	"deliverymanagement/internal/repo"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupAuthRouter() (*gin.Engine, *repo.InMemoryUserRepo) {
	repo := repo.NewInMemoryUserRepo()
	h := &AuthHandler{Users: repo}
	r := gin.Default()
	r.POST("/api/auth/register", h.Register)
	r.POST("/api/auth/login", h.Login)
	return r, repo
}

func TestRegisterAndLogin(t *testing.T) {
	r, _ := setupAuthRouter()

	// Register
	body := map[string]string{"email": "a@b.com", "password": "pass"}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Duplicate
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewReader(b))
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 400, w2.Code)

	// Login success
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewReader(b))
	r.ServeHTTP(w3, req3)
	assert.Equal(t, 200, w3.Code)
	var resp map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp["token"])

	// Login fail (wrong password)
	body2 := map[string]string{"email": "a@b.com", "password": "wrong"}
	b2, _ := json.Marshal(body2)
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewReader(b2))
	r.ServeHTTP(w4, req4)
	assert.Equal(t, 401, w4.Code)
}
