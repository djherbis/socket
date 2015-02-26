package socket

import "sync"

type PacketHandler interface {
	OnPacket(Packet)
}

type EventHandler interface {
	On(string, interface{}) error
	Handle(string, PacketHandler)
	PacketHandler
}

type handler struct {
	mu     sync.RWMutex
	events map[string]PacketHandler
}

func newHandler() EventHandler {
	return &handler{
		events: make(map[string]PacketHandler),
	}
}

func (h *handler) Handle(event string, c PacketHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events[event] = c
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
