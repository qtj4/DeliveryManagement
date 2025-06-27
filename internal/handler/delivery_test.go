package handler

import (
	"bytes"
	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
	"encoding/json"
	"fmt"
	"mime/multipart"
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

func makeDispatcherJWT(userID uint) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    "dispatcher",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	t, _ := token.SignedString(testSecret)
	return t
}

func makeCourierJWT(userID uint) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    "courier",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	t, _ := token.SignedString(testSecret)
	return t
}

func makeWarehouseJWT(userID uint) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    "warehouse",
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

func setupDeliveryRouterWithAssign() (*gin.Engine, *repo.InMemoryDeliveryRepo) {
	repo := repo.NewInMemoryDeliveryRepo()
	h := &DeliveryHandler{Deliveries: repo}
	r := gin.Default()
	deliveries := r.Group("/api/deliveries")
	deliveries.Use(JWTAuthMiddleware(testSecret))
	deliveries.POST("", h.CreateDelivery)
	deliveries.GET(":id", h.GetDelivery)
	deliveries.POST(":id/assign", DispatcherOnly(), h.AssignDelivery)
	return r, repo
}

func setupScanRouter() (*gin.Engine, *repo.InMemoryScanEventRepo) {
	repo := repo.NewInMemoryScanEventRepo()
	h := &ScanEventHandler{ScanEvents: repo}
	r := gin.Default()
	r.POST("/api/scan", JWTAuthMiddleware(testSecret), CourierOrWarehouseOnly(), h.CreateScanEvent)
	return r, repo
}

func setupDamageReportRouter() (*gin.Engine, *repo.InMemoryDamageReportRepo) {
	repo := repo.NewInMemoryDamageReportRepo()
	h := &DamageReportHandler{DamageReports: repo}
	r := gin.Default()
	r.POST("/api/damage-report", JWTAuthMiddleware(testSecret), CourierOrWarehouseOnly(), h.CreateDamageReport)
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

func TestAssignDeliveryRoleEnforcement(t *testing.T) {
	r, _ := setupDeliveryRouterWithAssign()
	jwtDispatcher := makeDispatcherJWT(1)
	jwtClient := makeJWT(2)

	// Create a delivery as dispatcher (could be any role)
	body := map[string]string{"from_address": "A", "to_address": "B"}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/deliveries", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+jwtDispatcher)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var created model.Delivery
	json.Unmarshal(w.Body.Bytes(), &created)

	// Dispatcher can assign
	assignBody := map[string]uint{"courier_id": 123}
	ab, _ := json.Marshal(assignBody)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", fmt.Sprintf("/api/deliveries/%d/assign", created.ID), bytes.NewReader(ab))
	req2.Header.Set("Authorization", "Bearer "+jwtDispatcher)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)
	var assigned model.Delivery
	json.Unmarshal(w2.Body.Bytes(), &assigned)
	assert.Equal(t, "ASSIGNED", assigned.Status)

	// Non-dispatcher cannot assign
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", fmt.Sprintf("/api/deliveries/%d/assign", created.ID), bytes.NewReader(ab))
	req3.Header.Set("Authorization", "Bearer "+jwtClient)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, 403, w3.Code)
}

func TestScanEventHandlers(t *testing.T) {
	r, repo := setupScanRouter()
	jwtCourier := makeCourierJWT(1)
	jwtWarehouse := makeWarehouseJWT(2)
	jwtClient := makeJWT(3)

	// Valid scan by courier
	body := map[string]interface{}{"delivery_id": 42, "event_type": "IN", "location": "Almaty Hub 3"}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/scan", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+jwtCourier)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var event model.ScanEvent
	json.Unmarshal(w.Body.Bytes(), &event)
	assert.Equal(t, uint(42), event.DeliveryID)
	assert.Equal(t, "IN", event.EventType)
	assert.Equal(t, "Almaty Hub 3", event.Location)
	assert.NotZero(t, event.Timestamp)

	// Valid scan by warehouse
	body2 := map[string]interface{}{"delivery_id": 43, "event_type": "OUT", "location": "Almaty Hub 3"}
	b2, _ := json.Marshal(body2)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/scan", bytes.NewReader(b2))
	req2.Header.Set("Authorization", "Bearer "+jwtWarehouse)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	// Invalid event_type
	body3 := map[string]interface{}{"delivery_id": 44, "event_type": "BAD", "location": "Almaty Hub 3"}
	b3, _ := json.Marshal(body3)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/api/scan", bytes.NewReader(b3))
	req3.Header.Set("Authorization", "Bearer "+jwtCourier)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, 400, w3.Code)

	// Missing location
	body4 := map[string]interface{}{"delivery_id": 45, "event_type": "IN"}
	b4, _ := json.Marshal(body4)
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("POST", "/api/scan", bytes.NewReader(b4))
	req4.Header.Set("Authorization", "Bearer "+jwtCourier)
	r.ServeHTTP(w4, req4)
	assert.Equal(t, 400, w4.Code)

	// Non-courier/warehouse forbidden
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest("POST", "/api/scan", bytes.NewReader(b))
	req5.Header.Set("Authorization", "Bearer "+jwtClient)
	r.ServeHTTP(w5, req5)
	assert.Equal(t, 403, w5.Code)

	// Tracking log append
	events := repo.ListScanEvents(42)
	assert.Len(t, events, 1)
	assert.Equal(t, "IN", events[0].EventType)
}

func TestDamageReportHandlers(t *testing.T) {
	r, repo := setupDamageReportRouter()
	jwtCourier := makeCourierJWT(1)
	jwtWarehouse := makeWarehouseJWT(2)
	jwtClient := makeJWT(3)

	// Valid report with file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("delivery_id", "42")
	_ = writer.WriteField("type", "box damaged")
	_ = writer.WriteField("description", "Corner crushed")
	fw, _ := writer.CreateFormFile("photo", "photo.jpg")
	fw.Write([]byte("imagedata"))
	writer.Close()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/damage-report", body)
	jwtWarehouse = makeWarehouseJWT(42)
	req.Header.Set("Authorization", "Bearer "+jwtWarehouse)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var report model.DamageReport
	json.Unmarshal(w.Body.Bytes(), &report)
	assert.Equal(t, uint(42), report.DeliveryID)
	assert.Equal(t, "box damaged", report.Type)
	assert.Equal(t, "Corner crushed", report.Description)
	assert.NotEmpty(t, report.PhotoPath)
	assert.NotZero(t, report.Timestamp)

	// Valid report without file (should fail, photo required)
	body2 := &bytes.Buffer{}
	writer2 := multipart.NewWriter(body2)
	_ = writer2.WriteField("delivery_id", "56")
	_ = writer2.WriteField("type", "torn label")
	_ = writer2.WriteField("description", "Label unreadable")
	writer2.Close()
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/damage-report", body2)
	req2.Header.Set("Authorization", "Bearer "+jwtWarehouse)
	req2.Header.Set("Content-Type", writer2.FormDataContentType())
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 400, w2.Code)

	// Missing fields
	body3 := &bytes.Buffer{}
	writer3 := multipart.NewWriter(body3)
	_ = writer3.WriteField("type", "broken seal")
	writer3.Close()
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/api/damage-report", body3)
	req3.Header.Set("Authorization", "Bearer "+jwtCourier)
	req3.Header.Set("Content-Type", writer3.FormDataContentType())
	r.ServeHTTP(w3, req3)
	assert.Equal(t, 400, w3.Code)

	// Non-courier/warehouse forbidden
	body4 := &bytes.Buffer{}
	writer4 := multipart.NewWriter(body4)
	_ = writer4.WriteField("delivery_id", "57")
	_ = writer4.WriteField("type", "broken seal")
	_ = writer4.WriteField("description", "Seal broken")
	writer4.Close()
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("POST", "/api/damage-report", body4)
	req4.Header.Set("Authorization", "Bearer "+jwtClient)
	req4.Header.Set("Content-Type", writer4.FormDataContentType())
	r.ServeHTTP(w4, req4)
	assert.Equal(t, 403, w4.Code)

	// Repo contains reports
	reports := repo.ListDamageReports(42)
	assert.Len(t, reports, 1)
	assert.Equal(t, "box damaged", reports[0].Type)
}
