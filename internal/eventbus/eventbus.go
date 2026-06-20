// Package eventbus is the coordinator's in-process pub/sub fan-out (ADR 0026).
// The scheduler and run supervisors publish run/state events; the HTTP SSE
// endpoint subscribes. It is non-blocking by design: a slow subscriber is
// dropped events rather than allowed to stall publishers or the scheduler loop.
package eventbus

import "sync"

// Event is a coordinator event delivered to subscribers. Seq is a monotonic
// hub-assigned id stamped at publish, so SSE can expose a stable Last-Event-ID
// offset (replay is not buffered in v1).
type Event struct {
	Seq      int64          `json:"seq"`
	Type     string         `json:"type"`
	RunID    string         `json:"run_id,omitempty"`
	TicketID string         `json:"ticket_id,omitempty"`
	Message  string         `json:"message,omitempty"`
	Payload  map[string]any `json:"payload,omitempty"`
}

// Hub fans published events out to all current subscribers.
type Hub struct {
	mu     sync.Mutex
	nextID int
	seq    int64
	subs   map[int]chan Event
	buffer int
	closed bool
}

// New returns a hub whose subscriber channels buffer up to bufferSize events
// before the hub starts dropping for that subscriber. A non-positive size uses a
// small default.
func New(bufferSize int) *Hub {
	if bufferSize <= 0 {
		bufferSize = 16
	}
	return &Hub{subs: map[int]chan Event{}, buffer: bufferSize}
}

// Subscribe registers a subscriber and returns its event channel plus an
// unsubscribe function. Unsubscribing is idempotent and closes the channel.
func (h *Hub) Subscribe() (<-chan Event, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	id := h.nextID
	h.nextID++
	ch := make(chan Event, h.buffer)
	if h.closed {
		close(ch)
		return ch, func() {}
	}
	h.subs[id] = ch
	var once sync.Once
	return ch, func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			if c, ok := h.subs[id]; ok {
				delete(h.subs, id)
				close(c)
			}
		})
	}
}

// Publish delivers ev to every subscriber without blocking: a subscriber whose
// buffer is full misses the event.
func (h *Hub) Publish(ev Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seq++
	ev.Seq = h.seq
	for _, ch := range h.subs {
		select {
		case ch <- ev:
		default: // drop for this slow subscriber
		}
	}
}

// Close closes all subscriber channels and rejects future subscriptions.
func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.closed = true
	for id, ch := range h.subs {
		delete(h.subs, id)
		close(ch)
	}
}
