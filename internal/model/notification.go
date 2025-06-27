package model

import (
	"time"
)

type Notification struct {
	ID        uint64                 `json:"id"`
	UserID    uint64                 `json:"user_id"`
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	IsRead    bool                   `json:"is_read"`
	CreatedAt time.Time              `json:"created_at"`
}
