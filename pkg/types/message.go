package types

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// Message struct
type Message struct {
	Data      []byte         `json:"data"`
	Others    cmap.CustomMap `json:"others"`
	Timestamp time.Time      `json:"timestamp"`
}

func (m *Message) ToString() string {
	data := ""
	if len(m.Data) > 0 {
		data = string(m.Data)
	}
	return fmt.Sprintf("{data:%s, others:%+v, timestamp:%v", data, m.Others, m.Timestamp)
}

// NewMessage returns a brand new message
func NewMessage(data []byte) *Message {
	return &Message{
		Timestamp: time.Now(),
		Data:      data,
		Others:    make(map[string]interface{}),
	}
}
