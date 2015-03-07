package socket

import (
	"errors"
	"sync"
)

var ErrNotSocketFunc = errors.New("connection/disconnection must take fn of type func(Socket)")

type Namespace interface {
	Name() string
	To(string) Emitter
	EventHandler
	Emitter
}

type namespace struct {
	mu           sync.RWMutex
	global       Room
	rooms        *roomSet
	onConnect    func(Socket)
	onDisconnect func(Socket)
	Handler
}

func newNamespace(name string) *namespace {
	return &namespace{
		global:  newRoom(name),
		rooms:   newRoomSet(),
		Handler: newHandler(),
	}
}

func (ns *namespace) Name() string { return ns.global.Name() }

func (ns *namespace) To(room string) Emitter {
	return ns.rooms.room(room)
}

func (ns *namespace) On(event string, fn interface{}) error {
	switch event {
	case Connection:
		sfn, ok := fn.(func(Socket))

		if !ok {
			return ErrNotSocketFunc
		}

		ns.mu.Lock()
		ns.onConnect = sfn
		ns.mu.Unlock()

	case Disconnection:
		sfn, ok := fn.(func(Socket))

		if !ok {
			return ErrNotSocketFunc
		}

		ns.mu.Lock()
		ns.onDisconnect = sfn
		ns.mu.Unlock()

	default:
		return ns.Handler.On(event, fn)
	}

	return nil
}

func (ns *namespace) Emit(event string, args ...interface{}) error {
	return ns.global.Emit(event, args...)
}

func (ns *namespace) addSocket(so Socket) {
	ns.mu.RLock()
	fn := ns.onConnect
	ns.mu.RUnlock()

	ns.global.Join(so)
	so.Join(so.Id())

	if fn != nil {
		fn(so)
	}
}

func (ns *namespace) removeSocket(so Socket) {
	ns.mu.RLock()
	fn := ns.onDisconnect
	ns.mu.RUnlock()

	ns.global.Leave(so)

	if fn != nil {
		fn(so)
	}
}
