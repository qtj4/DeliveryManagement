package handler

import (
	"bytes"
	"deliverymanagement/internal/repo"
	"deliverymanagement/pkg/rabbitmq"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupDeliveryRouterWithPublisher(pub rabbitmq.Publisher) (*gin.Engine, *repo.InMemoryDeliveryRepo, *rabbitmq.FakePublisher) {
	repo := repo.NewInMemoryDeliveryRepo()
	h := &DeliveryHandler{Deliveries: repo, Publisher: pub}
	r := gin.Default()
	deliveries := r.Group("/api/deliveries")
	deliveries.Use(JWTAuthMiddleware([]byte("supersecret")))
	deliveries.POST("", h.CreateDelivery)
	return r, repo, pub.(*rabbitmq.FakePublisher)
}

func TestDeliveryCreatedPublishesEvent(t *testing.T) {
	pub := &rabbitmq.FakePublisher{}
	r, _, fake := setupDeliveryRouterWithPublisher(pub)
	body := map[string]string{"from_address": "A", "to_address": "B"}
	b, _ := json.Marshal(body)
	jwt := makeJWT(1)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/deliveries", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+jwt) // valid JWT for user_id=1
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, fake.Messages)
	assert.Equal(t, "email.queue", fake.Messages[0].Queue)
	m := fake.Messages[0].Body.(map[string]interface{})
	assert.Equal(t, "delivery.created", m["event"])
}
