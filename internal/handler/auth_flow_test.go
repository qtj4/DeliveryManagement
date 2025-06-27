package handler

import (
	"deliverymanagement/internal/model"
	"deliverymanagement/internal/repo"
	"testing"
	"time"
)

func TestGenerateToken_UniqueAndLength(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		tok := generateToken()
		if len(tok) < 16 {
			t.Errorf("token too short: %s", tok)
		}
		if tokens[tok] {
			t.Errorf("duplicate token: %s", tok)
		}
		tokens[tok] = true
	}
}

func TestSetResetTokenAndExpiry(t *testing.T) {
	repo := repo.NewInMemoryUserRepo()
	u := &model.User{Email: "a@b.com"}
	repo.CreateUser(u)
	exp := time.Now().Add(1 * time.Hour)
	repo.SetResetToken(u.Email, "tok123", exp)
	user, _ := repo.FindUserByEmail(u.Email)
	if user.ResetToken != "tok123" {
		t.Errorf("expected reset token to be set")
	}
	if !user.ResetTokenExpiry.Equal(exp) {
		t.Errorf("expected expiry to match")
	}
}

func TestVerifyUser(t *testing.T) {
	repo := repo.NewInMemoryUserRepo()
	u := &model.User{Email: "a@b.com", VerificationToken: "tok"}
	repo.CreateUser(u)
	repo.VerifyUser(u.Email)
	user, _ := repo.FindUserByEmail(u.Email)
	if !user.IsVerified {
		t.Errorf("user should be verified")
	}
	if user.VerificationToken != "" {
		t.Errorf("verification token should be cleared")
	}
}
