package socket

import "sync"

type roomSet struct {
	mu    sync.Mutex
	rooms map[string]Room
}

func newRoomSet() *roomSet {
	return &roomSet{rooms: make(map[string]Room)}
}

func (r *roomSet) room(name string) Room {
	r.mu.Lock()
	defer r.mu.Unlock()
	if room, ok := r.rooms[name]; ok {
		return room
	}
	room := newRoom(name)
	r.rooms[name] = room
	return room
}

func (r *roomSet) Join(name string, so Socket) {
	r.mu.Lock()
	defer r.mu.Unlock()
	room, ok := r.rooms[name]
	if !ok {
		room = newRoom(name)
		r.rooms[name] = room
	}
	room.Join(so)
}

func (r *roomSet) Leave(name string, so Socket) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if room, ok := r.rooms[name]; ok {
		room.Leave(so)
		if room.Size() == 0 {
			delete(r.rooms, name)
		}
	}
}
