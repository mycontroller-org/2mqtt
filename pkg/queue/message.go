package queue

import (
	"sync"

	"github.com/mycontroller-org/2mqtt/pkg/types"
	"go.uber.org/zap"
)

// MessageQueue struct
type MessageQueue struct {
	ID       string
	Limit    int
	Messages []*types.Message
	mutex    *sync.RWMutex
}

// New creates a brand new MessageQueue
func New(ID string, limit int) *MessageQueue {
	return &MessageQueue{
		ID:       ID,
		Messages: make([]*types.Message, 0),
		Limit:    limit,
		mutex:    &sync.RWMutex{},
	}
}

// Add a message into the queue
func (q *MessageQueue) Add(msg *types.Message) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.Messages = append(q.Messages, msg)
	if len(q.Messages) > q.Limit {
		zap.L().Warn("dropping a message", zap.String("id", q.ID))
	}
}

// Get a message from the queue, if empty returns nil
func (q *MessageQueue) Get() *types.Message {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.Messages) > 0 {
		message := q.Messages[0]
		q.Messages = q.Messages[1:]
		return message
	}
	return nil
}
