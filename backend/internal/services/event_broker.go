// ============================================================
// FILE: internal/services/event_broker.go
// WHAT IT IS:     In-memory pub/sub for real-time SSE events
// WHY IT EXISTS:  The dashboard subscribes to a long-lived SSE
//                 connection so it learns about new contact-form
//                 submissions the instant they happen, instead of
//                 waiting for the next 30s poll.
// PRODUCERS:      message_handler.Create (publishes message.created)
// CONSUMERS:      sse_handler (one subscriber per connected dashboard)
// LAST UPDATED:   2026-05-10 — initial creation
// ============================================================

package services

import (
	"sync"
	"sync/atomic"
)

// Event is a single broker payload. The Type is what frontends switch on
// (e.g. "message.created"); Data is the JSON payload that goes after the
// SSE "data:" line. Keep Data small — every connected dashboard receives a copy.
type Event struct {
	Type string
	Data interface{}
}

// EventBroker is a process-local pub/sub. It is intentionally NOT durable
// — if the server restarts, in-flight events are lost and dashboards just
// reconnect and keep going. That's fine for "ping me when something changed"
// notifications; persistence is handled by the database row itself.
type EventBroker struct {
	mu          sync.RWMutex
	subscribers map[uint64]chan Event
	nextID      uint64
}

func NewEventBroker() *EventBroker {
	return &EventBroker{
		subscribers: make(map[uint64]chan Event),
	}
}

// Subscribe returns a buffered channel that receives every event published
// from now on, plus an unsubscribe func the caller MUST call when done so
// the channel doesn't leak. Buffer size 8 is enough for typical bursts;
// slow consumers drop excess events rather than blocking the publisher.
func (b *EventBroker) Subscribe() (<-chan Event, func()) {
	id := atomic.AddUint64(&b.nextID, 1)
	ch := make(chan Event, 8)
	b.mu.Lock()
	b.subscribers[id] = ch
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		if c, ok := b.subscribers[id]; ok {
			delete(b.subscribers, id)
			close(c)
		}
		b.mu.Unlock()
	}
}

// Publish fans the event out to every current subscriber. Non-blocking on
// each delivery: if a subscriber's buffer is full (slow / hung), we skip it
// rather than wedge the whole broker. The event is dropped for that one
// subscriber, which is acceptable for "refresh now" notifications.
func (b *EventBroker) Publish(evt Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers {
		select {
		case ch <- evt:
		default:
			// subscriber's buffer full — drop, don't block
		}
	}
}
