package model

import "time"

type Delivery struct {
	ID          uint
	FromAddress string
	ToAddress   string
	Status      string
	CreatedAt   time.Time
	DeliveredAt time.Time
	CourierID   uint
}
