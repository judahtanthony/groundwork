package eventbus

import (
	"sync"
	"testing"
)

func TestSubscribeReceives(t *testing.T) {
	h := New(4)
	ch, cancel := h.Subscribe()
	defer cancel()

	h.Publish(Event{Type: "run.started", RunID: "R-0001"})
	got := <-ch
	if got.Type != "run.started" || got.RunID != "R-0001" {
		t.Fatalf("got %+v", got)
	}
}

func TestUnsubscribeClosesAndStops(t *testing.T) {
	h := New(4)
	ch, cancel := h.Subscribe()
	cancel()
	if _, ok := <-ch; ok {
		t.Error("channel should be closed after cancel")
	}
	// Publishing after unsubscribe must not panic.
	h.Publish(Event{Type: "x"})
	cancel() // idempotent
}

func TestSlowSubscriberDropsNotBlocks(t *testing.T) {
	h := New(1)
	_, cancel := h.Subscribe() // never drained
	defer cancel()
	// More publishes than the buffer must not block.
	for i := 0; i < 100; i++ {
		h.Publish(Event{Type: "spam"})
	}
}

func TestConcurrentPubSub(t *testing.T) {
	h := New(64)
	var wg sync.WaitGroup

	// Subscribers churn.
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch, cancel := h.Subscribe()
			defer cancel()
			for range ch {
			}
		}()
	}
	// Publishers churn.
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				h.Publish(Event{Type: "tick"})
			}
		}()
	}
	// Close after publishers finish to release subscribers.
	go func() {
		for i := 0; i < 800; i++ {
			h.Publish(Event{Type: "warm"})
		}
		h.Close()
	}()
	wg.Wait()
}
