package socket

import "net/http"

type Socket struct {
	id string
	ns Namespace
	EventHandler
	t Transport
}

func newSocket(ns Namespace, p Packet) *Socket {
	return &Socket{
		id:           p.Socket(),
		EventHandler: newHandler(),
		ns:           ns,
		t:            p.Transport(),
	}
}

func (s *Socket) Id() string { return s.id }

func (s *Socket) Join(room string) {
	s.ns.Room(room).Join(s)
}

// TODO drop room on empty
func (s *Socket) Leave(room string) {
	s.ns.Room(room).Leave(s)
}

// TODO drop room on empty
func (s *Socket) To(room string) Emitter {
	return s.ns.Room(room)
}

func (s *Socket) Emit(event string, args ...interface{}) error {
	return s.t.Send(s.ns.Name(), s.Id(), event, args...)
}

func (s *Socket) Request() *http.Request {
	return s.t.Request()
}
