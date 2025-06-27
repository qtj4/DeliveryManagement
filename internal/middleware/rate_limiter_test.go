package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimiterMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := performRequest(r, "GET", "/test", nil)
	assert.Equal(t, 200, w.Code)

	// Simulate burst: 101 requests
	var lastCode int
	for i := 0; i < 101; i++ {
		w = performRequest(r, "GET", "/test", nil)
		lastCode = w.Code
	}
	assert.Equal(t, 429, lastCode)

	// Wait for rate limit window to reset
	time.Sleep(time.Minute)
	w = performRequest(r, "GET", "/test", nil)
	assert.Equal(t, 200, w.Code)
}

func performRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
