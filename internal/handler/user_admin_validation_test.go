package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"deliverymanagement/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateUserValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &UserAdminHandler{Users: newMockUserRepo()}
	r.POST("/api/admin/users", h.CreateUser)

	// Invalid email and short password
	body := map[string]interface{}{
		"email":    "notanemail",
		"name":     "",
		"role":     "invalidrole",
		"password": "123",
	}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/admin/users", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	assert.Contains(t, w.Body.String(), "Email")
	assert.Contains(t, w.Body.String(), "Role")
	assert.Contains(t, w.Body.String(), "Password")
}

// newMockUserRepo returns a dummy UserRepository for testing
func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{}
}

type mockUserRepo struct{}

func (m *mockUserRepo) CreateUser(u *model.User) error                             { return nil }
func (m *mockUserRepo) ListUsers() ([]*model.User, error)                          { return nil, nil }
func (m *mockUserRepo) UpdateUser(u *model.User) error                             { return nil }
func (m *mockUserRepo) DeleteUser(email string) error                              { return nil }
func (m *mockUserRepo) FindUserByEmail(email string) (*model.User, error)          { return nil, nil }
func (m *mockUserRepo) SetResetToken(email, token string, expires time.Time) error { return nil }
func (m *mockUserRepo) SetVerificationToken(email, token string) error             { return nil }
func (m *mockUserRepo) VerifyUser(email string) error                              { return nil }
