package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
)

type fakePublisher struct {
	Messages []map[string]interface{}
}

func (f *fakePublisher) Publish(queue string, body interface{}) error {
	f.Messages = append(f.Messages, body.(map[string]interface{}))
	return nil
}
func (f *fakePublisher) Close() error { return nil }

func TestExportDeliveries_SyncCSV(t *testing.T) {
	repo := repo.NewInMemoryDeliveryRepo()
	repo.CreateDelivery(&model.Delivery{ID: 1, FromAddress: "A", ToAddress: "B", Status: "CREATED"})
	h := &DeliveryHandler{Deliveries: repo}
	r := gin.Default()
	r.GET("/export", h.ExportDeliveries)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/export?format=csv", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "FromAddress,ToAddress,Status")
}

func TestExportDeliveries_AsyncJob(t *testing.T) {
	repo := repo.NewInMemoryDeliveryRepo()
	// Simulate large dataset
	for i := 0; i < 2001; i++ {
		repo.CreateDelivery(&model.Delivery{ID: uint(i), FromAddress: "A", ToAddress: "B", Status: "CREATED"})
	}
	pub := &fakePublisher{}
	h := &DeliveryHandler{Deliveries: repo, Publisher: pub}
	r := gin.Default()
	r.GET("/export", h.ExportDeliveries)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/export?format=csv&email=test@example.com", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 202, w.Code)
	assert.NotEmpty(t, pub.Messages)
	msg := pub.Messages[0]
	assert.True(t, strings.HasPrefix(msg["job_id"].(string), "job-"))
	assert.Equal(t, "csv", msg["format"])
	assert.Equal(t, "test@example.com", msg["email"])
}
