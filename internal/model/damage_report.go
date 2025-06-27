package model

import "time"

type DamageReport struct {
	ID          uint
	DeliveryID  uint
	Type        string
	Description string
	PhotoPath   string // relative path to uploaded photo
	PhotoSize   int64  // file size in bytes
	PhotoMime   string // mime type
	Timestamp   time.Time
}
