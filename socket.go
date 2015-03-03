package socket

import (
	"net/http"
	"sync"
)

type Socket interface {
	Id() string
	Join(string)
	Leave(string)
	Rooms() []string
	To(string) Emitter
	EventHandler
	Emitter
}

type socket struct {
	mu sync.RWMutex
	id string
	ns Namespace
	EventHandler
	t     Transport
	rooms map[string]struct{}
}

func newSocket(ns Namespace, p Packet) Socket {
	return &socket{
		id:           p.Socket(),
		EventHandler: newHandler(),
		ns:           ns,
		t:            p.Transport(),
		rooms:        make(map[string]struct{}),
	}
}

func (s *socket) Id() string { return s.id }

func (s *socket) Join(room string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rooms[room] = struct{}{}
	s.ns.Room(room).Join(s)
}

// TODO drop room on empty
func (s *socket) Leave(room string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rooms, room)
	s.ns.Room(room).Leave(s)
}

// TODO drop room on empty
func (s *socket) To(room string) Emitter {
	return s.ns.Room(room)
}

func (s *socket) Emit(event string, args ...interface{}) error {
	return s.t.Send(s.ns.Name(), s.Id(), event, args...)
}

func (s *socket) Request() *http.Request {
	return s.t.Request()
}

func (s *socket) Rooms() (rooms []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for room, _ := range s.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
