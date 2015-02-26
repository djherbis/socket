package socket

import (
	"errors"
	"sync"
)

type Namespace interface {
	Name() string
	Of(name string) Namespace
	Room(string) Room
	EventHandler
	Emitter
}

type namespace struct {
	mu        sync.RWMutex
	name      string
	sockets   map[string]*Socket
	subspaces map[string]*namespace
	rooms     map[string]Room
	EventHandler
}

func newNamespace(name string) *namespace {
	ns := &namespace{
		name:         name,
		sockets:      make(map[string]*Socket),
		subspaces:    make(map[string]*namespace),
		rooms:        make(map[string]Room),
		EventHandler: newHandler(),
	}

	ns.EventHandler.Handle(CONNECTION, &connecter{namespace: ns})
	ns.EventHandler.Handle(DISCONNECTION, &disconnecter{namespace: ns})

	return ns
}

func (ns *namespace) Room(name string) Room {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	if room, ok := ns.rooms[name]; ok {
		return room
	}
	room := newRoom(name)
	ns.rooms[name] = room
	return room
}

func (ns *namespace) Name() string { return ns.name }

func (ns *namespace) Of(name string) Namespace {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	if nsp, ok := ns.subspaces[name]; ok {
		return nsp
	}
	subns := newNamespace(name)
	ns.subspaces[name] = subns
	return subns
}

var ErrNotSocketFunc = errors.New("connection/disconnection must take fn of type func(*Socket)")

func (ns *namespace) On(event string, fn interface{}) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	switch event {
	case CONNECTION:
		sfn, ok := fn.(func(*Socket))

		if !ok {
			return ErrNotSocketFunc
		}

		ns.EventHandler.Handle(event, &connecter{
			namespace: ns,
			fn:        sfn,
		})

	case DISCONNECTION:
		sfn, ok := fn.(func(*Socket))

		if !ok {
			return ErrNotSocketFunc
		}

		ns.EventHandler.Handle(event, &disconnecter{
			namespace: ns,
			fn:        sfn,
		})

	default:
		return ns.EventHandler.On(event, fn)
	}

	return nil
}

func (ns *namespace) route(name string) (subns *namespace, ok bool) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	if name == "" || name == "/" {
		return ns, true
	}
	subns, ok = ns.subspaces[name]
	return subns, ok
}

func (ns *namespace) OnPacket(p Packet) {
	if nsp, ok := ns.route(p.Namespace()); ok {
		nsp.EventHandler.OnPacket(p)
		nsp.Room(p.Socket()).OnPacket(p)
	}
}

func (ns *namespace) Emit(event string, args ...interface{}) {
	ns.Room("").Emit(event, args...)
}

type connecter struct {
	*namespace
	fn func(*Socket)
}

func (c *connecter) OnPacket(p Packet) {
	so := newSocket(c.namespace, p)
	c.namespace.mu.Lock()
	c.namespace.sockets[so.Id()] = so
	c.namespace.mu.Unlock()
	so.Join(so.Id())
	so.Join("")
	if c.fn != nil {
		c.fn(so)
	}
}

type disconnecter struct {
	*namespace
	fn func(*Socket)
}

func (c *disconnecter) OnPacket(p Packet) {
	c.namespace.mu.Lock()
	so, ok := c.namespace.sockets[p.Socket()]
	if ok {
		delete(c.namespace.sockets, p.Socket())
	}
	c.namespace.mu.Unlock()

	if ok {
		for _, room := range c.namespace.rooms {
			room.Leave(so)
		}

		if c.fn != nil {
			c.fn(so)
		}
	}
}
