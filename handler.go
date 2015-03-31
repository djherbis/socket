package socket

import "sync"

// PacketHandler responds to a Packet
type PacketHandler interface {
	OnPacket(Packet)
}

// EventHandler registers a function to be run when an event is received.
// The arguments to the function will be unmarshalled from the
// javascript objects emitted by the client-side socket.
type EventHandler interface {
	On(event string, fn interface{}) error
}

// Handler is both a PacketHandler and a EventHandler
type Handler interface {
	EventHandler
	PacketHandler
}

type handler struct {
	mu     sync.RWMutex
	events map[string]PacketHandler
}

func newHandler() Handler {
	return &handler{
		events: make(map[string]PacketHandler),
	}
}

func (h *handler) On(event string, fn interface{}) (err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events[event], err = newPacketHandler(fn)
	return err
}

func (h *handler) OnPacket(p Packet) {
	h.mu.RLock()
	c, ok := h.events[p.Event()]
	h.mu.RUnlock()

	if ok {
		c.OnPacket(p)
	}
}
