package socket

import (
	"errors"
	"net/http"
	"sync"
)

type Emitter interface {
	Emit(string, ...interface{}) error
}

type Socket interface {
	Id() string
	Namespace() string
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
	Handler
	t            Transport
	rooms        map[string]struct{}
	onDisconnect func()
}

func newSocket(ns Namespace, p Packet) *socket {
	return &socket{
		id:      p.Socket(),
		Handler: newHandler(),
		ns:      ns,
		t:       p.Transport(),
		rooms:   make(map[string]struct{}),
	}
}

func (s *socket) Id() string { return s.id }

func (s *socket) Namespace() string { return s.ns.Name() }

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

func (s *socket) On(event string, fn interface{}) error {
	switch event {
	case Disconnect:
		sfn, ok := fn.(func())
		if !ok {
			return errors.New("Disconnect takes a func of type func()")
		}

		s.mu.Lock()
		s.onDisconnect = sfn
		s.mu.Unlock()
		return nil

	default:
		return s.Handler.On(event, fn)
	}
}

func (s *socket) disconnect() {
	s.mu.RLock()
	for room, _ := range s.rooms {
		s.ns.Room(room).Leave(s)
		delete(s.rooms, room)
	}
	fn := s.onDisconnect
	s.mu.RUnlock()

	if fn != nil {
		fn()
	}
}
