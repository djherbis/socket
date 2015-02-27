package socket

import (
	"net/http"
	"sync"
)

type Socket struct {
	mu sync.RWMutex
	id string
	ns Namespace
	EventHandler
	t     Transport
	rooms map[string]struct{}
}

func newSocket(ns Namespace, p Packet) *Socket {
	return &Socket{
		id:           p.Socket(),
		EventHandler: newHandler(),
		ns:           ns,
		t:            p.Transport(),
		rooms:        make(map[string]struct{}),
	}
}

func (s *Socket) Id() string { return s.id }

func (s *Socket) Join(room string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rooms[room] = struct{}{}
	s.ns.Room(room).Join(s)
}

// TODO drop room on empty
func (s *Socket) Leave(room string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rooms, room)
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

func (s *Socket) Rooms() (rooms []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for room, _ := range s.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
