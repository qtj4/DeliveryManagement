package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"deliverymanagement/internal/handler"
	"deliverymanagement/internal/repo"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestWebSocketScanBroadcast(t *testing.T) {
	gin.SetMode(gin.TestMode)
	scanRepo := repo.NewInMemoryScanEventRepo()
	h := &handler.ScanEventHandler{ScanEvents: scanRepo}
	r := gin.Default()
	r.GET("/ws/track/:deliveryId", handler.TrackWSHandler)
	r.POST("/api/scan", h.CreateScanEvent)

	ts := httptest.NewServer(r)
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	u.Scheme = "ws"
	u.Path = "/ws/track/42"
	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	defer ws.Close()

	// Trigger scan event
	scan := map[string]interface{}{
		"delivery_id": 42,
		"event_type":  "IN",
		"location":    "Warehouse",
	}
	b, _ := json.Marshal(scan)
	resp, err := http.Post(ts.URL+"/api/scan", "application/json", bytes.NewReader(b))
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Should receive broadcast
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg map[string]interface{}
	err = ws.ReadJSON(&msg)
	assert.NoError(t, err)
	assert.Equal(t, "scan.updated", msg["event"])
	assert.EqualValues(t, 42, msg["delivery_id"])
}
