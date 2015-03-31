package socket

import (
	"net/http"
	"sync"
)

type Emitter interface {
	Emit(string, ...interface{}) error
}

type Socket interface {
	ClientSocket
	Join(string)
	Leave(string)
	Rooms() []string
	To(string) Emitter
	Request() *http.Request
}

type socket struct {
	mu sync.RWMutex
	*clientSocket
	ns    *namespace
	rooms map[string]struct{}
}

func newSocket(ns *namespace, id string, t Transport) *socket {
	return &socket{
		clientSocket: newClientSocket(id, ns.Name(), t),
		ns:           ns,
		rooms:        make(map[string]struct{}),
	}
}

func (s *socket) Join(room string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rooms[room] = struct{}{}
	s.ns.rooms.Join(room, s)
}

func (s *socket) Leave(room string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rooms, room)
	s.ns.rooms.Leave(room, s)
}

func (s *socket) To(room string) Emitter {
	return s.ns.To(room)
}

func (s *socket) Request() *http.Request {
	return s.clientSocket.Request()
}

func (s *socket) Rooms() (rooms []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for room, _ := range s.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

func (s *socket) disconnect() {
	s.clientSocket.disconnect()

	s.mu.RLock()
	defer s.mu.RUnlock()

	for room, _ := range s.rooms {
		s.ns.rooms.Leave(room, s)
		delete(s.rooms, room)
	}
}
