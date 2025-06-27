package model

import "time"

type User struct {
	ID                uint
	Email             string
	Name              string
	Role              string
	PasswordHash      string
	IsVerified        bool
	VerificationToken string
	ResetToken        string
	ResetTokenExpiry  time.Time
}
