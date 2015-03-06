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
	name         string
	rooms        *roomSet
	onConnect    func(Socket)
	onDisconnect func(Socket)
	Handler
}

func newNamespace(name string) *namespace {
	return &namespace{
		name:    name,
		rooms:   newRoomSet(),
		Handler: newHandler(),
	}
}

func (ns *namespace) Name() string { return ns.name }

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
	return ns.To("").Emit(event, args...)
}

func (ns *namespace) addSocket(so Socket) {
	ns.mu.RLock()
	fn := ns.onConnect
	ns.mu.RUnlock()

	if fn != nil {
		so.Join("")
		so.Join(so.Id())
		fn(so)
	}
}

func (ns *namespace) removeSocket(so Socket) {
	ns.mu.RLock()
	fn := ns.onDisconnect
	ns.mu.RUnlock()

	if fn != nil {
		fn(so)
	}
}
