package services

import "sync"

type EventHub struct {
	mu   sync.Mutex
	subs map[chan string]int64
}

func NewEventHub() *EventHub {
	return &EventHub{subs: map[chan string]int64{}}
}

func (h *EventHub) Subscribe(klienID int64) chan string {
	ch := make(chan string, 16)
	h.mu.Lock()
	h.subs[ch] = klienID
	h.mu.Unlock()
	return ch
}

func (h *EventHub) Unsubscribe(ch chan string) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
}

// Publish hanya sampai ke pelanggan stream milik klien yang sama.
func (h *EventHub) Publish(klienID int64, events ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ev := range events {
		for ch, kid := range h.subs {
			if kid != klienID {
				continue
			}
			select {
			case ch <- ev:
			default: // klien lambat dilewati, ia refetch di event berikutnya
			}
		}
	}
}
