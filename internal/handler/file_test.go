package handler

import (
	"bytes"
	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupFileTestRouter() (*gin.Engine, *repo.InMemoryDamageReportRepo) {
	repo := repo.NewInMemoryDamageReportRepo()
	h := &FileHandler{DamageReports: repo}
	r := gin.Default()
	r.GET("/files/:filename", func(c *gin.Context) {
		// Simulate JWT middleware
		c.Set("user_id", uint(42)) // user_id matches DeliveryID
		c.Set("role", "admin")     // use admin for guaranteed access
		h.ServeFile(c)
	})
	return r, repo
}

func TestFileAccess_RoleBased(t *testing.T) {
	r, repo := gin.Default(), repo.NewInMemoryDamageReportRepo()
	h := &FileHandler{DamageReports: repo}
	// Simulate a real upload to get the actual PhotoPath
	dh := &DamageReportHandler{DamageReports: repo}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("delivery_id", "42")
	_ = writer.WriteField("type", "box damaged")
	_ = writer.WriteField("description", "Corner crushed")
	fw, _ := writer.CreateFormFile("photo", "photo.jpg")
	fw.Write([]byte("imagedata"))
	writer.Close()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = reqWithMultipartBody("/api/damage-report", body, writer.FormDataContentType())
	dh.CreateDamageReport(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var report model.DamageReport
	json.Unmarshal(w.Body.Bytes(), &report)

	// Allowed: user_id matches DeliveryID
	r.GET("/files/:filename", func(c *gin.Context) {
		c.Set("user_id", uint(42))
		c.Set("role", "courier")
		h.ServeFile(c)
	})
	rec := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/files/"+report.PhotoPath, nil)
	r.ServeHTTP(rec, req2)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Forbidden: user_id does not match
	r2 := gin.Default()
	r2.GET("/files/:filename", func(c *gin.Context) {
		c.Set("user_id", uint(99))
		c.Set("role", "courier")
		h.ServeFile(c)
	})
	rec2 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/files/"+report.PhotoPath, nil)
	r2.ServeHTTP(rec2, req3)
	assert.Equal(t, http.StatusForbidden, rec2.Code)
}

func reqWithMultipartBody(url string, body *bytes.Buffer, contentType string) *http.Request {
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", contentType)
	return req
}
