package model

import "time"

type ScanEvent struct {
	ID         uint
	DeliveryID uint
	EventType  string // "IN" or "OUT"
	Location   string
	Timestamp  time.Time
}
