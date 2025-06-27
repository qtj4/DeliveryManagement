package handler

import (
	"bytes"
	"deliverymanagement/internal/repo"
	"deliverymanagement/pkg/rabbitmq"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupDamageReportRouterWithPublisher(pub rabbitmq.Publisher) (*gin.Engine, *repo.InMemoryDamageReportRepo, *rabbitmq.FakePublisher) {
	repo := repo.NewInMemoryDamageReportRepo()
	h := &DamageReportHandler{DamageReports: repo, Publisher: pub}
	r := gin.Default()
	r.POST("/api/damage-report", h.CreateDamageReport)
	return r, repo, pub.(*rabbitmq.FakePublisher)
}

func TestDamageReportPublishesEvent(t *testing.T) {
	pub := &rabbitmq.FakePublisher{}
	r, _, fake := setupDamageReportRouterWithPublisher(pub)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.WriteField("delivery_id", "1")
	w.WriteField("type", "broken")
	w.WriteField("description", "test")
	// Add a dummy file
	fileWriter, _ := w.CreateFormFile("photo", "test.jpg")
	fileWriter.Write([]byte("dummydata"))
	w.Close()

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/damage-report", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, fake.Messages)
	assert.Equal(t, "email.queue", fake.Messages[0].Queue)
	m := fake.Messages[0].Body.(map[string]interface{})
	assert.Equal(t, "damage.reported", m["event"])
}
