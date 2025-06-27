package model

import "time"

type DamageReport struct {
	ID          uint
	DeliveryID  uint
	Type        string
	Description string
	PhotoPath   string // relative path to uploaded photo
	Timestamp   time.Time
}
