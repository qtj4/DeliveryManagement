package handler

import (
	"bytes"
	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

var testSecret = []byte("supersecret")

func makeJWT(userID uint) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    "client",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	t, _ := token.SignedString(testSecret)
	return t
}

func setupDeliveryRouter() (*gin.Engine, *repo.InMemoryDeliveryRepo) {
	repo := repo.NewInMemoryDeliveryRepo()
	h := &DeliveryHandler{Deliveries: repo}
	r := gin.Default()
	deliveries := r.Group("/api/deliveries")
	deliveries.Use(JWTAuthMiddleware(testSecret))
	deliveries.POST("", h.CreateDelivery)
	deliveries.GET(":id", h.GetDelivery)
	return r, repo
}

func TestDeliveryHandlers(t *testing.T) {
	r, _ := setupDeliveryRouter()
	jwt := makeJWT(1)

	// Unauthorized create
	body := map[string]string{"from_address": "A", "to_address": "B"}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/deliveries", bytes.NewReader(b))
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)

	// Authorized create
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/deliveries", bytes.NewReader(b))
	req2.Header.Set("Authorization", "Bearer "+jwt)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)
	var created model.Delivery
	json.Unmarshal(w2.Body.Bytes(), &created)
	assert.Equal(t, "CREATED", created.Status)
	assert.Equal(t, "A", created.FromAddress)
	assert.Equal(t, "B", created.ToAddress)

	// Fetch by ID
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", fmt.Sprintf("/api/deliveries/%d", created.ID), nil)
	req3.Header.Set("Authorization", "Bearer "+jwt)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, 200, w3.Code)
	var fetched model.Delivery
	json.Unmarshal(w3.Body.Bytes(), &fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, created.Status, fetched.Status)

	// Not found
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("GET", "/api/deliveries/9999", nil)
	req4.Header.Set("Authorization", "Bearer "+jwt)
	r.ServeHTTP(w4, req4)
	assert.Equal(t, 404, w4.Code)
}
