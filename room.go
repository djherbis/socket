package socket

import "sync"

// Room is a collection of Sockets
type Room interface {

	// Name returns the room name
	Name() string

	// Size returns the # of sockets in the room
	Size() int

	// Join adds a socket to the Room
	Join(so Socket)

	// Leave removes a socket from the Room
	Leave(so Socket)

	// Emit sends to all members of the room
	Emitter
}

type room struct {
	sync.RWMutex
	name    string
	sockets map[Socket]struct{}
}

func newRoom(name string) Room {
	return &room{
		name:    name,
		sockets: make(map[Socket]struct{}),
	}
}

func (r *room) Name() string {
	return r.name
}

func (r *room) Join(so Socket) {
	r.Lock()
	defer r.Unlock()
	r.sockets[so] = struct{}{}
}

func (r *room) Leave(so Socket) {
	r.Lock()
	defer r.Unlock()
	delete(r.sockets, so)
}

func (r *room) Size() int {
	r.RLock()
	defer r.RUnlock()
	return len(r.sockets)
}

func (r *room) getSockets() []Socket {
	r.RLock()
	defer r.RUnlock()
	sockets := make([]Socket, len(r.sockets))
	i := 0
	for so, _ := range r.sockets {
		sockets[i] = so
		i++
	}
	return sockets
}

func (r *room) Emit(event string, args ...interface{}) (err error) {
	for _, so := range r.getSockets() {
		if er := so.Emit(event, args...); er != nil {
			err = err
		}
	}
	return err
}
