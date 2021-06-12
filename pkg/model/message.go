package model

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Message struct
type Message struct {
	Data      interface{}    `json:"data"`
	Others    cmap.CustomMap `json:"others"`
	Timestamp time.Time      `json:"timestamp"`
}

// NewMessage returns a brand new message
func NewMessage(data interface{}) *Message {
	return &Message{
		Timestamp: time.Now(),
		Data:      data,
		Others:    make(map[string]interface{}),
	}
}
