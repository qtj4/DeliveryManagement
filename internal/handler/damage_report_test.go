package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"deliverymanagement/internal/repo"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateDamageReport_UploadValidation(t *testing.T) {
	repo := repo.NewInMemoryDamageReportRepo()
	h := &DamageReportHandler{DamageReports: repo}
	r := gin.Default()
	r.POST("/api/damage-report", h.CreateDamageReport)

	// Valid JPG
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.WriteField("delivery_id", "1")
	w.WriteField("type", "broken")
	w.WriteField("description", "desc")
	fw, _ := w.CreateFormFile("photo", "test.jpg")
	fw.Write([]byte("dummydata"))
	w.Close()

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/damage-report", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Invalid file type
	b.Reset()
	w = multipart.NewWriter(&b)
	w.WriteField("delivery_id", "1")
	fw, _ = w.CreateFormFile("photo", "test.txt")
	fw.Write([]byte("dummydata"))
	w.Close()

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/damage-report", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// Too large
	b.Reset()
	w = multipart.NewWriter(&b)
	w.WriteField("delivery_id", "1")
	fw, _ = w.CreateFormFile("photo", "test.jpg")
	fw.Write(make([]byte, 6*1024*1024)) // 6MB
	w.Close()

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/damage-report", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
